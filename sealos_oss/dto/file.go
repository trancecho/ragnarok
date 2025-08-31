package dto

type AddFileReq struct {
	Name     string `json:"name" form:"name"`
	FolderId uint   `json:"folder_id" form:"folder_id"`
}

type UpdateFileReq struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type DeleteFileReq struct {
	Id uint `json:"id"`
}

type GetFilesReq struct {
	FolderId uint   `json:"folder_id" form:"folder_id"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	SortBy   string `json:"sortby" form:"sortby"`
}

type FileResponse struct {
	Id   uint   `json:"file_id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Path string `json:"path"`
}
