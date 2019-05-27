package types

type BasicAuth struct {
	Secret string `json:"secret"`
}

type ImageMetadata struct {
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"` // After encoding
}

type ImageUploadRequest struct {
	Auth       BasicAuth     `json:"auth"`
	Metadata   ImageMetadata `json:"metadata"`
	Payload    string        `json:"payload"`
	Checksum   string        `json:"checksum"` // SHA256 before encoding
	AccessType string        `json:"access_type"`
}

// Auth optional for public images only.
type ImageListRequest struct {
	Auth       BasicAuth `json:"auth,omitempty"`
	AccessType string    `json:"access_type"`
}

type ImageListResponse struct {
	Images []ImageMetadata `json:"images"`
}
