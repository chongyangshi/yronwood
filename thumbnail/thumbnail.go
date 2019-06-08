package thumbnail

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/monzo/slog"
	"github.com/nfnt/resize"

	"github.com/icydoge/yronwood/config"
)

const thumbnailWidth = 200

func GetThumbnailForImage(ctx context.Context, fileName, storagePath, accessType string) ([]byte, error) {
	thumbnailPath := config.ConfigStorageDirectoryThumbnail
	if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
		slog.Info(ctx, "Thumbnail directory %s does not exist, attempting to create it", thumbnailPath)
		mkdirErr := os.Mkdir(config.ConfigStorageDirectoryThumbnail, 0755)
		if mkdirErr != nil {
			slog.Error(ctx, "Could not create non-existing storage directory %s: %v", thumbnailPath, err)
			return nil, mkdirErr
		}
	}

	thumbnailFilePath := path.Join(thumbnailPath, getThumbnailFileName(fileName, accessType))
	if _, err := os.Stat(thumbnailFilePath); err == nil {
		thumbnail, err := ioutil.ReadFile(thumbnailFilePath)
		if err != nil {
			slog.Debug(ctx, "Could not read thumbnail file %s: %v", thumbnailFilePath, err)
			return nil, err
		}
		// Found thumbnail already processed, return it.
		return thumbnail, nil
	} else if !os.IsNotExist(err) {
		// Unknown thumbnail file read error, bail
		slog.Debug(ctx, "Could not check if thumbnail file %s exists: %v", thumbnailFilePath, err)
		return nil, err
	}

	// File not thumbnailed before, we need to process and store the thumbnail.
	filePath := path.Join(storagePath, fileName)
	if _, err := os.Stat(filePath); err != nil {
		slog.Debug(ctx, "Cannot read image %s storage path %s, not making thumbnail", fileName, storagePath)
		return nil, err
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		slog.Debug(ctx, "Could not read file %s when making thumbnail: %v", filePath, err)
		return nil, err
	}

	img, err := decodeImage(fileName, file)
	if err != nil {
		slog.Debug(ctx, "Could not decode image %s when making thumbnail: %v", filePath, err)
		return nil, err
	}
	if img == nil {
		slog.Debug(ctx, "Could not select image %s decode method when making thumbnail: %v", filePath, err)
		return nil, err
	}

	thumbnail := resize.Thumbnail(thumbnailWidth, 0, img, resize.Lanczos3)
	encodedThumbnail, err := encodeImage(fileName, thumbnail)
	if err != nil {
		slog.Debug(ctx, "Could not encode thumbnail of image %s: %v", filePath, err)
		return nil, err
	}

	if err = ioutil.WriteFile(getThumbnailFileName(fileName, accessType), encodedThumbnail, 0644); err != nil {
		slog.Debug(ctx, "Could not write thumbnail of image %s: %v", filePath, err)
		return nil, err
	}

	return encodedThumbnail, nil
}

func getThumbnailFileName(fileName, accessType string) string {
	fileNameComponents := strings.Split(fileName, ".")
	if len(fileNameComponents) < 2 {
		return fileName
	}

	name := strings.Join(fileNameComponents[0:len(fileNameComponents)-1], ".")
	extension := fileNameComponents[len(fileNameComponents)-1]
	thumbnailFileName := fmt.Sprintf("%s_%s.%s", name, "thumb", extension)

	return thumbnailFileName
}

func decodeImage(fileName string, filePayload []byte) (image.Image, error) {
	fileNameComponents := strings.Split(fileName, ".")
	if len(fileNameComponents) < 2 {
		return nil, nil
	}
	extension := fileNameComponents[len(fileNameComponents)-1]

	switch strings.ToLower(extension) {
	case "jpg", "jpeg":
		return jpeg.Decode(bytes.NewReader(filePayload))
	case "png":
		return png.Decode(bytes.NewReader(filePayload))
	case "gif":
		return gif.Decode(bytes.NewReader(filePayload))
	}

	return nil, nil
}

func encodeImage(fileName string, img image.Image) ([]byte, error) {
	fileNameComponents := strings.Split(fileName, ".")
	if len(fileNameComponents) < 2 {
		return nil, nil
	}
	extension := fileNameComponents[len(fileNameComponents)-1]

	var buf = bytes.Buffer{}
	imageWriter := bufio.NewWriter(&buf)

	var encodeErr error
	switch strings.ToLower(extension) {
	case "jpg", "jpeg":
		encodeErr = jpeg.Encode(imageWriter, img, nil)
	case "png":
		encodeErr = png.Encode(imageWriter, img)
	case "gif":
		encodeErr = gif.Encode(imageWriter, img, nil)
	}

	flushErr := imageWriter.Flush()
	if encodeErr != nil || flushErr != nil {
		return nil, encodeErr
	}

	imageReader := bufio.NewReader(&buf)
	return ioutil.ReadAll(imageReader)
}
