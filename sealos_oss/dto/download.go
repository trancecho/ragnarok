package dto

type DownloadFilesReq struct {
	FileId uint `json:"file_id" form:"file_id"`
}

type DownloadResponse struct {
	DownloadUrl string `json:"downloadUrl"`
	ExpiresAt   string `json:"expiresAt"`
	FileInfo    struct {
		Id      uint   `json:"id"`
		Name    string `json:"name"`
		Size    int64  `json:"size"`
		Hotness int64  `json:"hotness"`
	} `json:"fileInfo"`
}

type PreviewResponse struct {
	PreviewUrl string `json:"previewUrl"`
	ExpiresAt  string `json:"expiresAt"`
	FileInfo   struct {
		Id          uint   `json:"id"`
		Name        string `json:"name"`
		Size        int64  `json:"size"`
		ContentType string `json:"contentType"`
	} `json:"fileInfo"`
}
