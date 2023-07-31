package config

import (
	"os"
	"strings"
)

var (
	ConfigListenAddr                = getConfigFromOSEnv("YRONWOOD_LISTEN_ADDR", ":8080")
	ConfigIndexRedirect             = getConfigFromOSEnv("YRONWOOD_INDEX_REDIRECT", "https://images.chongya.ng")
	ConfigStorageDirectoryPublic    = getConfigFromOSEnv("YRONWOOD_STORAGE_DIRECTORY_PUBLIC", "/images/uploads/public")
	ConfigStorageDirectoryUnlisted  = getConfigFromOSEnv("YRONWOOD_STORAGE_DIRECTORY_UNLISTED", "/images/uploads/big")
	ConfigStorageDirectoryPrivate   = getConfigFromOSEnv("YRONWOOD_STORAGE_DIRECTORY_PRIVATE", "/images/uploads/private")
	ConfigStorageDirectoryThumbnail = getConfigFromOSEnv("YRONWOOD_STORAGE_DIRECTORY_THUMBNAIL", "/images/uploads/thumbnail")
	ConfigAccessTypePublic          = getConfigFromOSEnv("YRONWOOD_ACCESS_TYPE_PUBLIC", "public")
	ConfigAccessTypeUnlisted        = getConfigFromOSEnv("YRONWOOD_ACCESS_TYPE_UNLISTED", "big")
	ConfigAccessTypePrivate         = getConfigFromOSEnv("YRONWOOD_ACCESS_TYPE_PRIVATE", "private")
	ConfigMaxFileSize               = getConfigFromOSEnv("YRONWOOD_MAX_FILE_SIZE", "25165824")  // 24MB
	ConfigMaxFileNameSize           = getConfigFromOSEnv("YRONWOOD_MAX_FILE_NAME_SIZE", "1024") // GCP max
	ConfigPermittedExtensions       = getConfigFromOSEnv("YRONWOOD_PERMITTED_EXTENSIONS", "jpeg|jpg|png|gif")
	ConfigAuthenticationSigningKey  = getConfigFromOSEnv("YRONWOOD_AUTHENTICATION_SIGHNING_KEY", "unit_test")
	ConfigAuthenticationBasicSecret = getConfigFromOSEnv("YRONWOOD_AUTHENTICATION_BASIC_SECRET", "unit_test")
	ConfigAuthenticationBasicSalt   = getConfigFromOSEnv("YRONWOOD_AUTHENTICATION_BASIC_SALT", "unit_test")
	ConfigCORSAllowedOrigin         = getConfigFromOSEnv("YRONWOOD_CORS_ALLOWED_ORIGIN", "https://images.chongya.ng")
)

// This is intended to run inside Kubernetes as a pod, so we just set service Configurations from deployment Configuration.
func getConfigFromOSEnv(key, defaultValue string) string {
	envValue := os.Getenv(key)
	if envValue != "" {
		return envValue
	}

	return defaultValue
}

// FileExtensionToContentType returns the appropriate HTTP content type for a given extension.
func FileExtensionToContentType(extension string) string {
	switch strings.ToLower(extension) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	}

	return "application/octet-stream"
}
