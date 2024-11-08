package types

type AuthenticateRequest struct {
	Secret string `json:"secret"`
}

type AuthenticateResponse struct {
	Token string `json:"token"`
}

type ImageMetadata struct {
	FileName   string   `json:"file_name"`
	Tags       []string `json:"tags"`
	AccessPath string   `json:"access_path"`
	Uploaded   string   `json:"uploaded"`
	ImageToken string   `json:"image_token"` // Pre-signed read access token for private images only
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
	Token      string   `json:"token"`
	AccessType string   `json:"access_type"`
	Page       int      `json:"page"`
	Tags       []string `json:"tags"`
}

type ImageListResponse struct {
	Images         []ImageMetadata `json:"images"`
	PagesAvailable bool            `json:"next_page"`
}

type ImageDeleteRequest struct {
	Token      string `json:"token"`
	FileName   string `json:"file_name"`
	AccessType string `json:"access_type"`
}
