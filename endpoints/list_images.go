package endpoints

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"

	"github.com/chongyangshi/yronwood/auth"
	"github.com/chongyangshi/yronwood/config"
	"github.com/chongyangshi/yronwood/types"
)

const (
	defaultPagingStart = 0
	defaultPagingCount = 21
	pageLimit          = 105

	imageTokenValidity = time.Duration(time.Hour * 12)
)

type imageMetadata struct {
	FileName   string
	AccessPath string
	Uploaded   time.Time

	// Pre-signed read access token for private images only
	ImageToken string
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
		authSuccess, err := auth.VerifyAdminToken(body.Token)
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
			imageMeta := imageMetadata{
				FileName:   pathFile.Name(),
				AccessPath: accessType,
				Uploaded:   pathFile.ModTime(),
			}

			if accessType == config.ConfigAccessTypePrivate {
				imageToken, err := auth.SignImageToken(
					imageTokenValidity,
					fmt.Sprintf("%s/%s", config.ConfigAccessTypePrivate, pathFile.Name()),
				)

				if err != nil {
					slog.Error(req, "Error pre-signing image %s in directory %s: %v", pathFile.Name(), storagePath, err)
					return typhon.Response{Error: terrors.InternalService("", "Error encountered listing images", nil)}
				}

				imageMeta.ImageToken = imageToken
			}

			files = append(files, imageMeta)
		}
	}

	var images = []imageMetadata{}
	for _, file := range files {
		images = append(images, file)
	}

	// Most recent first. This application deals with small number of files, so backend
	// always lists all files of the access directory presently.
	sort.Slice(images, func(i, j int) bool {
		return images[i].Uploaded.After(images[j].Uploaded)
	})

	page := body.Page
	if page > pageLimit {
		// Hard stack limit on runtimes not optimising tail recursions.
		page = pageLimit
	}
	start, end := boundPaging(body.Page, len(images))
	return req.Response(types.ImageListResponse{Images: internalMetadataToResponseList(images[start:end])})
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

func boundPaging(page, available int) (int, int) {
	if page <= 0 {
		return defaultPagingStart, boundMin(defaultPagingCount, available)
	}

	if page*defaultPagingCount <= available {
		return (page - 1) * defaultPagingCount, page * defaultPagingCount
	}

	// Tail recursive, sadly golang likely still doesn't optimise for it.
	return boundPaging(page-1, available)
}

func boundMin(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
