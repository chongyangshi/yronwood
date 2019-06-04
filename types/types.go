package types

type AuthenticateRequest struct {
	Secret string `json:"secret"`
}

type AuthenticateResponse struct {
	Token string `json:"token"`
}

type ImageMetadata struct {
	FileName   string `json:"file_name"`
	AccessPath string `json:"access_path"`
	Uploaded   string `json:"uploaded"`
}

type ImageUploadRequest struct {
	Token      string        `json:"token"`
	Metadata   ImageMetadata `json:"metadata"`
	Payload    string        `json:"payload"`
	Checksum   string        `json:"checksum"` // SHA256 after encoding
	AccessType string        `json:"access_type"`
}

// Auth optional for public images only.
type ImageListRequest struct {
	Token      string `json:"token"`
	AccessType string `json:"access_type"`
	Page       int    `json:"page"`
}

type ImageListResponse struct {
	Images []ImageMetadata `json:"images"`
}
