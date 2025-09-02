package dto

import "github.com/trancecho/ragnarok/sealos_oss/entity"

type AddFolderReq struct {
	Name           string `json:"name"`
	ParentFolderId uint   `json:"parent_folder_id"`
}

type GetSonFoldersReq struct {
	Id       uint   `json:"id" form:"id"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	SortBy   string `json:"sortby" form:"sortby"`
}

type UpdateFolderReq struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type DeleteFolderReq struct {
	Id uint `json:"id"`
}

type ListFolderContentsReq struct {
	FolderId uint `json:"folder_id" form:"folder_id"`
}

type FolderOperationResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	FolderId  uint   `json:"folder_id,omitempty"`
	FolderKey string `json:"folder_key,omitempty"`
}

type FolderContents struct {
	Files   []string        `json:"files"`
	Folders []entity.Folder `json:"folders"`
}
