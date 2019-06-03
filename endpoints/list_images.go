package endpoints

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"

	"github.com/icydoge/yronwood/auth"
	"github.com/icydoge/yronwood/config"
	"github.com/icydoge/yronwood/types"
)

type imageMetadata struct {
	FileName   string
	AccessPath string
	Uploaded   time.Time
}

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

	validAccessType, _ := validateAccessType(body.AccessType)
	if !validAccessType {
		return typhon.Response{Error: terrors.BadRequest("invalid_access_type", "Access type specified is invalid", nil)}
	}

	if body.AccessType != config.ConfigAccessTypePublic {
		authSuccess, err := auth.VerifyToken(body.Token)
		if err != nil {
			slog.Error(req, "Error authenticating: %v", err)
			return typhon.Response{Error: terrors.InternalService("", "Error encountered authenticating you", nil)}
		}
		if !authSuccess {
			if body.Token == "" {
				return typhon.Response{Error: terrors.Unauthorized("", "Authentication required", nil)}
			}
			return typhon.Response{Error: terrors.Forbidden("bad_access", "Unauthorized access for this access type", nil)}
		}
	}

	storagePaths := accessTypeToPaths(body.AccessType)
	files := []imageMetadata{}
	for accessType, storagePath := range storagePaths {
		pathFiles, err := ioutil.ReadDir(storagePath)
		if err != nil && !os.IsNotExist(err) {
			slog.Error(req, "Error listing images in directory %s: %v", storagePath, err)
			return typhon.Response{Error: terrors.InternalService("", "Error encountered listing images", nil)}
		}

		for _, pathFile := range pathFiles {
			files = append(files, imageMetadata{
				FileName:   pathFile.Name(),
				AccessPath: accessType,
				Uploaded:   pathFile.ModTime(),
			})
		}
	}

	var images = []imageMetadata{}
	for _, file := range files {
		images = append(images, file)
	}

	// Most recent first.
	sort.Slice(images, func(i, j int) bool {
		return images[i].Uploaded.After(images[j].Uploaded)
	})

	return req.Response(types.ImageListResponse{Images: internalMetadataToResponseList(images)})
}

func internalMetadataToResponseList(images []imageMetadata) []types.ImageMetadata {
	files := []types.ImageMetadata{}
	for _, image := range images {
		files = append(files, types.ImageMetadata{
			FileName:   image.FileName,
			AccessPath: image.AccessPath,
			Uploaded:   image.Uploaded.Format(time.RFC3339),
		})
	}

	return files
}
