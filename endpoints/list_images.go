package endpoints

import (
	"encoding/json"
	"io/ioutil"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"

	"github.com/icydoge/yronwood/config"
	"github.com/icydoge/yronwood/types"
)

func listImages(req typhon.Request) typhon.Response {
	imageListRequest, err := req.BodyBytes(true)
	if err != nil {
		slog.Error(req, "Error reading request body: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	body := types.ImageListRequest{}
	err = json.Unmarshal(imageListRequest, &body)
	if err != nil {
		slog.Error(req, "Error parsing request body: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	validAccessType, storagePath := validateAccessType(body.AccessType)
	if !validAccessType {
		return typhon.Response{Error: terrors.BadRequest("invalid_access_type", "Access type specified is invalid", nil)}
	}

	if body.AccessType != config.ConfigAccessTypePublic {
		authSuccess, err := doBasicAuth(body.Auth.Secret)
		if err != nil {
			slog.Error(req, "Error authenticating: %v", err)
			return typhon.Response{Error: terrors.InternalService("", "Error encountered authenticating you", nil)}
		}
		if !authSuccess {
			if body.Auth.Secret == "" {
				return typhon.Response{Error: terrors.Unauthorized("", "Authentication required", nil)}
			}
			return typhon.Response{Error: terrors.Unauthorized("bad_access", "Unauthorized access for this access type", nil)}
		}
	}

	files, err := ioutil.ReadDir(storagePath)
	if err != nil {
		slog.Error(req, "Error listing images in directory %s: %v", storagePath, err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered listing images", nil)}
	}

	var images = []types.ImageMetadata{}
	for _, f := range files {
		images = append(images, types.ImageMetadata{
			FileName: f.Name(),
		})
	}

	return req.Response(types.ImageListResponse{Images: images})
}
