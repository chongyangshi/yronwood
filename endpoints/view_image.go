package endpoints

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"

	"github.com/icydoge/yronwood/auth"
	"github.com/icydoge/yronwood/config"
)

func viewImage(req typhon.Request) typhon.Response {
	err := req.ParseForm()
	if err != nil {
		slog.Error(req, "Error processing query params: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	// For legacy compatible reasons, we process image url in a pseudo-static manner (hostname/accesstype/filename.jpg)
	success, accessType, fileName := processURI(req.URL.Path)
	if !success {
		return typhon.Response{Error: terrors.NotFound("not_found", "Requested image is not found", nil)}
	}

	// Auth optional for public images and unlisted images.
	authenticated, err := auth.VerifyToken(req.FormValue("token"))
	if err != nil {
		slog.Error(req, "Error authenticating client: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	if accessType == config.ConfigAccessTypePrivate && !authenticated {
		if req.FormValue("token") == "" {
			return typhon.Response{Error: terrors.Unauthorized("", "Authentication required", nil)}
		}
		return typhon.Response{Error: terrors.Forbidden("", "Authentication failure", nil)}
	}

	if !validateFilename(fileName) {
		return typhon.Response{Error: terrors.BadRequest("invalid_filename", fmt.Sprintf("File name %s is invalid", fileName), nil)}
	}

	imageBytes := readStoredImageByAccessType(req, fileName, accessType)
	if imageBytes != nil {
		response := typhon.NewResponse(req)
		response.Body = ioutil.NopCloser(bytes.NewReader(imageBytes))
		response.Header.Set("Content-Type", getContentTypeFromFilename(fileName))
		return response
	}

	return typhon.Response{Error: terrors.NotFound("not_found", fmt.Sprintf("Requested image %s is not found", fileName), nil)}
}

func readStoredImageByAccessType(ctx context.Context, fileName, accessType string) []byte {
	var imageBytes []byte
	switch accessType {
	case config.ConfigAccessTypePublic:
		imageBytes = readFile(ctx, config.ConfigStorageDirectoryPublic, fileName)
	case config.ConfigAccessTypeUnlisted:
		imageBytes = readFile(ctx, config.ConfigStorageDirectoryUnlisted, fileName)
	case config.ConfigAccessTypePrivate:
		imageBytes = readFile(ctx, config.ConfigStorageDirectoryPrivate, fileName)
	}

	return imageBytes
}

// For legacy-compatible reasons, we process image url in a pseudo-static manner (apihost/accesstype/filename.jpg)
// This function attempts to extract access type (public/unlisted) and file name from the URI.
func processURI(URI string) (bool, string, string) {
	trimmedURI := strings.Trim(URI, "/")
	splitURI := strings.Split(trimmedURI, "/")
	if len(splitURI) < 3 {
		return false, "", ""
	}

	return true, splitURI[1], splitURI[2]
}
