package dto

// FileInfo возвращается на клиент
type FileInfo struct {
	ID      int    `json:"id"`
	FileURL string `json:"file_url"`
}
