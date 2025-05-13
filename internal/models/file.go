package models

// DeleteFileRequest represents a request to delete a file
type DeleteFileRequest struct {
	FileURL string `json:"file_url" binding:"required"`
}
