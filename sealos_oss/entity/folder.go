package entity

import (
	"time"
)

type Folder struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Uid            uint      `json:"uid" form:"uid"`
	Name           string    `json:"name" form:"name"`
	ParentFolderId uint      `json:"parent_folder_id" form:"parent_folder_id"`
}
