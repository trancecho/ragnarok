package entity

import (
	"time"
)

type File struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Uid       uint      `json:"uid" form:"uid"`
	FolderId  uint      `json:"folder_id" form:"folder_id"`
	Name      string    `json:"name" form:"name"`
	MineType  string    `json:"mine_type" form:"mine_type"`
	Size      int64     `json:"size" form:"size"`
	Hotness   int64     `json:"hotness" form:"hotness"`
}
