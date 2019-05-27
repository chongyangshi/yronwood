package endpoints

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"

	"github.com/icydoge/yronwood/config"
)

var (
	permittedExtensions  = strings.Split(config.ConfigPermittedExtensions, "|")
	permittedComposition = regexp.MustCompile(`[a-zA-Z0-9]+`)
)

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

	inputSHA256 := base64.StdEncoding.EncodeToString(inputHash.Sum(nil))
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
		if extension == fileNameSplit[1] {
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

func validateAccessType(accessType string) (bool, string) {
	switch accessType {
	case config.ConfigAccessTypePublic:
		return true, config.ConfigStorageDirectoryPublic
	case config.ConfigAccessTypeUnlisted:
		return false, config.ConfigStorageDirectoryUnlisted
	case config.ConfigAccessTypePrivate:
		return false, config.ConfigStorageDirectoryPrivate
	}

	return false, ""
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
