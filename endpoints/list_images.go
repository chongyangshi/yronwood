package endpoints

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
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
	pagingCount        = 21
	imageTokenValidity = time.Duration(time.Hour * 12)
)

type imageMetadata struct {
	FileName   string
	Tags       []string
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

	// Prepare tags for filtering images within the current accessible paths,
	// if present.
	filterTags := map[string]bool{}
	for _, tag := range body.Tags {
		filterTags[tag] = true
	}

	storagePaths := accessTypeToPaths(body.AccessType)
	files := map[string]imageMetadata{}
	for accessType, storagePath := range storagePaths {
		pathFiles, err := ioutil.ReadDir(storagePath)
		if err != nil && !os.IsNotExist(err) {
			slog.Error(req, "Error listing images in directory %s: %v", storagePath, err)
			return typhon.Response{Error: terrors.InternalService("", "Error encountered listing images", nil)}
		}

		for _, pathFile := range pathFiles {
			fileName := pathFile.Name()

			// If any tags are present on the image, separate them out. This has no effect
			// for images without tags.
			var tags []string
			fileName, tags, err = decodeFileNameWithTags(fileName)
			if err != nil {
				slog.Error(req, "Error decoding tags for %s in directory %s: %v", fileName, storagePath, err)
				return typhon.Response{Error: terrors.InternalService("", "Error encountered listing images", nil)}
			}

			// If any tag filters are present, only images which match at least
			// one tag in the filter will be returned.
			if len(filterTags) != 0 {
				if len(tags) == 0 {
					continue
				}

				matched := false
				for _, tag := range tags {
					if _, found := filterTags[tag]; found {
						matched = true
					}
				}
				if !matched {
					continue
				}
			}

			imageMeta := imageMetadata{
				FileName:   fileName,
				Tags:       tags,
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

			// Images with tags will be processed multiple times due to presence of
			// symlinks for tagged file names. Overwrite correctly to include tag
			// information where applicable.
			if existing, exists := files[fileName]; exists {
				if len(existing.Tags) != 0 && len(imageMeta.Tags) == 0 {
					continue
				}
			}

			files[fileName] = imageMeta
		}
	}

	// Sort a copy
	var images = []imageMetadata{}
	for _, file := range files {
		images = append(images, file)
	}

	// Most recent first. This application deals with small number of files, so backend
	// always lists all files of the access directory presently.
	sort.Slice(images, func(i, j int) bool {
		return images[i].Uploaded.After(images[j].Uploaded)
	})

	start, end := boundPaging(body.Page, len(images))
	return req.Response(types.ImageListResponse{
		Images:         internalMetadataToResponseList(images[start:end]),
		PagesAvailable: end < len(images),
	})
}

func internalMetadataToResponseList(images []imageMetadata) []types.ImageMetadata {
	files := []types.ImageMetadata{}
	for _, image := range images {
		files = append(files, types.ImageMetadata{
			FileName:   image.FileName,
			Tags:       image.Tags,
			AccessPath: image.AccessPath,
			Uploaded:   image.Uploaded.Format(time.RFC3339),
			ImageToken: image.ImageToken,
		})
	}

	return files
}

func boundPaging(pageNumber, imagesAvailable int) (int, int) {
	pageNumberMax := math.Ceil(float64(imagesAvailable) / pagingCount)

	if pageNumber > int(pageNumberMax) {
		pageNumber = int(pageNumberMax)
	}

	if pageNumber < 1 {
		pageNumber = 1
	}

	imageNumberStart := (pageNumber - 1) * pagingCount
	imageNumberEnd := pageNumber * pagingCount

	if imageNumberEnd > imagesAvailable {
		imageNumberEnd = imagesAvailable
	}

	return imageNumberStart, imageNumberEnd
}
