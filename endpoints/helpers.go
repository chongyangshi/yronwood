package endpoints

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"

	"github.com/chongyangshi/yronwood/config"
)

const tagSeparator = "|"

var (
	permittedExtensions        = strings.Split(config.ConfigPermittedExtensions, "|")
	permittedComposition       = regexp.MustCompile(`[a-zA-Z0-9-_]+`)
	maxFileNameSize      int64 = 1024
)

func init() {
	maxFileNameSizeParsed, err := strconv.ParseInt(config.ConfigMaxFileNameSize, 10, 32)
	if err == nil {
		maxFileNameSize = maxFileNameSizeParsed
	}
}

func doBasicAuth(secret string) (bool, error) {
	if secret == "" {
		return false, nil
	}

	saltedInput := []byte(fmt.Sprintf("%s:%s", secret, config.ConfigAuthenticationBasicSalt))

	return validateChecksum(saltedInput, config.ConfigAuthenticationBasicSecret)
}

func validateChecksum(payload []byte, checksum string) (bool, error) {
	inputHash := sha256.New()
	_, err := inputHash.Write(payload)
	if err != nil {
		return false, terrors.Wrap(err, nil)
	}

	inputSHA256 := hex.EncodeToString(inputHash.Sum(nil))
	if inputSHA256 == checksum {
		return true, nil
	}

	return false, nil
}

func getContentTypeFromFilename(fileName string) string {
	fileNameSplit := strings.SplitN(fileName, ".", 2)
	if len(fileNameSplit) != 2 {
		return "application/octet-stream"
	}

	return config.FileExtensionToContentType(fileNameSplit[1])
}

func validateFilename(fileName string) bool {
	fileNameSplit := strings.SplitN(fileName, ".", 2)
	if len(fileNameSplit) != 2 {
		return false
	}

	validExtension := false
	for _, extension := range permittedExtensions {
		if extension == strings.ToLower(fileNameSplit[1]) {
			validExtension = true
			break
		}
	}
	if !validExtension {
		return false
	}

	// Filename must be alphanumeric only
	matchedIndices := permittedComposition.FindStringIndex(fileNameSplit[0])
	if len(matchedIndices) < 2 {
		return false
	}

	if matchedIndices[0] != 0 || matchedIndices[1] != len(fileNameSplit[0]) {
		return false
	}

	return true
}

// Converts a filename and optionally a list of tags into an encoded filename for storage.
// Returns an error if the file name is invalid.
func validateAndEncodeFileNameWithTags(fileName string, tags []string) (string, error) {
	if !validateFilename(fileName) {
		return "", terrors.BadRequest("invalid_filename", fmt.Sprintf("Filename %s is invalid", fileName), nil)
	}

	if len(tags) == 0 {
		return fileName, nil
	}

	encodedTags := make([]string, len(tags))
	for i, tag := range tags {
		if !permittedComposition.MatchString(tag) {
			return "", terrors.BadRequest("invalid_tag", fmt.Sprintf("File tag %s contains invalid characters", tag), nil)
		}
		encodedTags[i] = base64.StdEncoding.EncodeToString([]byte(tag))
	}

	storedName := fileName
	if len(encodedTags) > 0 {
		storedName = fmt.Sprintf("%s%s%s", strings.Join(encodedTags, tagSeparator), tagSeparator, storedName)
	}
	if int64(len(storedName)) > maxFileNameSize {
		return "", terrors.BadRequest("tags_too_long", fmt.Sprintf("Encoded file tags when combined with file name %s are longer than limit %d", storedName, maxFileNameSize), nil)
	}

	return storedName, nil
}

// Reverses the process of encodeFileNameWithTags and recover filename and any tags from storage.
// Returns an error if the any tags cannot be decoded or the encoded filename is invalid.
func decodeFileNameWithTags(fileNameWithTags string) (string, []string, error) {
	fileNameSplit := strings.Split(fileNameWithTags, tagSeparator)
	if len(fileNameSplit) < 1 {
		return "", nil, terrors.InternalService("invalid_filename", "Invalid file name split, this should never happen", nil)
	}
	if len(fileNameSplit) == 1 {
		// No tags exist, return file name
		return fileNameSplit[0], nil, nil
	}

	encodedTags := fileNameSplit[:len(fileNameSplit)-1]
	fileName := fileNameSplit[len(fileNameSplit)-1]
	tags := make([]string, len(encodedTags))
	for i, encodedTag := range encodedTags {
		tag, err := base64.StdEncoding.DecodeString(encodedTag)
		if err != nil {
			return "", nil, terrors.WrapWithCode(err, map[string]string{"encoded_tag": encodedTag}, "decoding_tag")
		}
		tags[i] = string(tag)
	}

	return fileName, tags, nil
}

func validateAccessType(accessType string) (bool, string) {
	switch accessType {
	case config.ConfigAccessTypePublic:
		return true, config.ConfigStorageDirectoryPublic
	case config.ConfigAccessTypeUnlisted:
		return true, config.ConfigStorageDirectoryUnlisted
	case config.ConfigAccessTypePrivate:
		return true, config.ConfigStorageDirectoryPrivate
	}

	return false, ""
}

// Private access type also gives access to public and unlisted images
func accessTypeToPaths(accessType string) map[string]string {
	switch accessType {
	case config.ConfigAccessTypePublic:
		return map[string]string{
			config.ConfigAccessTypePublic: config.ConfigStorageDirectoryPublic,
		}
	case config.ConfigAccessTypeUnlisted:
		return map[string]string{
			config.ConfigAccessTypeUnlisted: config.ConfigStorageDirectoryUnlisted,
		}
	case config.ConfigAccessTypePrivate:
		return map[string]string{
			config.ConfigAccessTypePublic:   config.ConfigStorageDirectoryPublic,
			config.ConfigAccessTypeUnlisted: config.ConfigStorageDirectoryUnlisted,
			config.ConfigAccessTypePrivate:  config.ConfigStorageDirectoryPrivate,
		}
	}

	return nil
}

// readFile queries directory for existence of file, and if exists, return content
// as bytes. The fileName MUST be validated by validateFilename() before passing in.
func readFile(ctx context.Context, storagePath, fileName string) []byte {
	filePath := path.Join(storagePath, fileName)
	if _, err := os.Stat(filePath); err != nil {
		return nil
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		slog.Debug(ctx, "Could not read file %s: %v", filePath, err)
		return nil
	}

	return file
}

// deleteFile queries directory for existence of file, and if exists, delete
// the file. If any symlinks with the same file name suffix exists, then
// delete the symlinks as well.
func deleteFile(ctx context.Context, storagePath, fileName string) error {
	filePath := path.Join(storagePath, fileName)
	if _, err := os.Stat(filePath); err != nil {
		slog.Error(ctx, "Could not find file %s to delete: %v", filePath, err)
		return err
	}

	err := os.Remove(filePath)
	if err != nil {
		slog.Error(ctx, "Could not delete file %s: %v", filePath, err)
		return err
	}

	pathFiles, err := ioutil.ReadDir(storagePath)
	if err != nil {
		slog.Error(ctx, "Could not list directory for file %s: %v", storagePath, err)
		return err
	}

	for _, file := range pathFiles {
		if !strings.HasSuffix(file.Name(), fileName) {
			continue
		}

		suffixFilePath := path.Join(storagePath, file.Name())
		lstat, err := os.Lstat(suffixFilePath)
		if err != nil {
			slog.Error(ctx, "Could not check suffix file %s: %v", suffixFilePath, err)
			return err
		}

		if lstat.Mode()&os.ModeSymlink == 0 {
			slog.Error(ctx, "%s is not a symlink as expected for %s", suffixFilePath, filePath)
			return fmt.Errorf("%s is not a symlink as expected for %s", suffixFilePath, filePath)
		}

		if err = os.Remove(suffixFilePath); err != nil {
			slog.Error(ctx, "Could not remove suffix symlink %s: %v", suffixFilePath, err)
			return err
		}
	}

	return nil
}
