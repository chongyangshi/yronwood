package endpoints

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"

	"github.com/chongyangshi/yronwood/auth"
	"github.com/chongyangshi/yronwood/config"
	"github.com/chongyangshi/yronwood/types"
)

var maxUploadSize int64

func init() {
	var err error
	maxUploadSize, err = strconv.ParseInt(config.ConfigMaxFileSize, 10, 32)
	if err != nil {
		maxUploadSize = 24 * 1024 * 1024
	}
}

func uploadImage(req typhon.Request) typhon.Response {
	imageUploadRequest, err := req.BodyBytes(true)
	if err != nil {
		slog.Error(req, "Error reading request body: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	body := types.ImageUploadRequest{}
	err = json.Unmarshal(imageUploadRequest, &body)
	if err != nil {
		slog.Error(req, "Error parsing request body: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	// Auth required for uploading images.
	authenticated, err := auth.VerifyAdminToken(body.Token)
	if err != nil {
		slog.Error(req, "Error authenticating client: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}
	if !authenticated {
		if body.Token == "" {
			return typhon.Response{Error: terrors.Unauthorized("", "Authentication required", nil)}
		}
		return typhon.Response{Error: terrors.Forbidden("", "Authentication failure", nil)}
	}

	if len(body.Payload) == 0 || len(body.Payload) > int(maxUploadSize) {
		return typhon.Response{Error: terrors.BadRequest("bad_file_size", "Content length of payload is too large", nil)}
	}

	if !validateFilename(body.Metadata.FileName) {
		slog.Error(req, "Invalid file name: %+v", err)
		return typhon.Response{Error: terrors.BadRequest("bad_file_name", "Invalid file name or extension specified", nil)}
	}

	taggedFileName, err := validateAndEncodeFileNameWithTags(body.Metadata.FileName, body.Metadata.Tags)
	if err != nil {
		slog.Error(req, "Invalid file name or tags: %+v", err)
		return typhon.Response{Error: terrors.BadRequest("bad_file_tags", "Invalid file tags specified", nil)}
	}

	validAccessType, storagePath := validateAccessType(body.AccessType)
	if !validAccessType {
		return typhon.Response{Error: terrors.BadRequest("bad_access_type", "Invalid file access type specified", nil)}
	}

	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		slog.Info(req, "Storage directory %s does not exist, attempting to create it", storagePath)
		mkdirErr := os.Mkdir(storagePath, 0755)
		if mkdirErr != nil {
			slog.Error(req, "Could not create non-existing storage directory %s: %v", storagePath, err)
			return typhon.Response{Error: terrors.InternalService("", "Error encountered retrieving file", nil)}
		}
	} else if err != nil {
		slog.Error(req, "Could not check if storage directory exists: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered retrieving file", nil)}
	}

	validChecksum, err := validateChecksum([]byte(body.Payload), body.Checksum)
	if err != nil || !validChecksum {
		return typhon.Response{Error: terrors.BadRequest("bad_payload", "Invalid payload, could not verify checksum", nil)}
	}

	decodedPayload, err := base64.StdEncoding.DecodeString(body.Payload)
	if err != nil {
		slog.Error(req, "Error decoding base64 payload: %v", err)
		return typhon.Response{Error: terrors.BadRequest("bad_payload", "Invalid payload, could not decode", nil)}
	}

	if readStoredImageByAccessType(req, body.Metadata.FileName, body.AccessType) != nil {
		return typhon.Response{Error: terrors.BadRequest("file_exists", "File with given name already exists", nil)}
	}

	// Upload the original file.
	filePath := path.Join(storagePath, body.Metadata.FileName)
	err = ioutil.WriteFile(filePath, decodedPayload, 0644)
	if err != nil {
		return typhon.Response{Error: terrors.InternalService("", fmt.Sprintf("Could not save file: %v", err), nil)}
	}

	// If tags present, create a symlink from the tagged file name to the original file.
	// This symlink is used to store the tags in the file system only and would never be
	// actually followed by Yronwood, only listed and filtered.
	if taggedFileName == body.Metadata.FileName {
		return req.Response(nil)
	}

	symlinkPath := path.Join(storagePath, taggedFileName)
	if err := os.Symlink(filePath, symlinkPath); err != nil {
		return typhon.Response{Error: terrors.InternalService("", fmt.Sprintf("Could not create symlink for tagged file: %v", err), nil)}
	}

	return req.Response(nil)
}
