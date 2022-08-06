package endpoints

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"

	"github.com/chongyangshi/yronwood/auth"
	"github.com/chongyangshi/yronwood/config"
	"github.com/chongyangshi/yronwood/types"
)

func deleteImage(req typhon.Request) typhon.Response {
	imageDeleteRequest, err := req.BodyBytes(true)
	if err != nil {
		slog.Error(req, "Error reading request body: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	body := types.ImageDeleteRequest{}
	err = json.Unmarshal(imageDeleteRequest, &body)
	if err != nil {
		slog.Error(req, "Error parsing request body: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	// Must be authenticated as admin user to delete images
	authenticated, err := auth.VerifyAdminToken(body.Token)
	if err != nil {
		slog.Error(req, "Error authenticating client: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	if !authenticated {
		if req.FormValue("token") == "" {
			return typhon.Response{Error: terrors.Unauthorized("", "Authentication required", nil)}
		}
		return typhon.Response{Error: terrors.Forbidden("", "Authentication failure", nil)}
	}

	if !validateFilename(body.FileName) {
		return typhon.Response{Error: terrors.BadRequest("invalid_filename", fmt.Sprintf("File name %s is invalid", body.FileName), nil)}
	}

	err = deleteStoredImageByAccessType(req, body.FileName, body.AccessType)
	if err != nil {
		slog.Error(req, "Could not delete file %s of type %s: %+v", body.FileName, body.AccessType, err)
		return typhon.Response{Error: err}
	}

	return req.Response(nil)
}

func deleteStoredImageByAccessType(ctx context.Context, fileName, accessType string) error {
	var err error
	switch accessType {
	case config.ConfigAccessTypePublic:
		err = deleteFile(ctx, config.ConfigStorageDirectoryPublic, fileName)
	case config.ConfigAccessTypeUnlisted:
		err = deleteFile(ctx, config.ConfigStorageDirectoryUnlisted, fileName)
	case config.ConfigAccessTypePrivate:
		err = deleteFile(ctx, config.ConfigStorageDirectoryPrivate, fileName)
	}

	return err
}
