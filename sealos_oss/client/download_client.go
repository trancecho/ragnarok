package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/trancecho/ragnarok/sealos_oss/dto"
	"github.com/trancecho/ragnarok/sealos_oss/entity"
	"gorm.io/gorm"
)

type DownloadClient struct {
	client *SealOSClient
}

func NewDownloadClient(client *SealOSClient) *DownloadClient {
	return &DownloadClient{client: client}
}

func (dc *DownloadClient) GenerateDownloadURL(uid uint, req *dto.DownloadFilesReq) (*dto.DownloadResponse, error) {
	// 获取文件信息
	var file entity.File
	if err := dc.client.db.First(&file, req.FileId).Error; err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	// 验证文件权限
	if file.Uid != uid {
		return nil, fmt.Errorf("无权限访问此文件")
	}

	// 生成完整文件路径
	objectKey, err := dc.getFullFilePath(req.FileId)
	if err != nil {
		return nil, fmt.Errorf("获取文件路径失败: %w", err)
	}

	// 更新文件热度
	if err := dc.client.db.Model(&file).
		Update("hotness", gorm.Expr("hotness + 1")).Error; err != nil {
		return nil, fmt.Errorf("更新文件热度失败: %w", err)
	}

	// 生成预签名URL
	url, err := dc.generatePresignedURL(objectKey)
	if err != nil {
		return nil, fmt.Errorf("生成下载链接失败: %w", err)
	}

	return &dto.DownloadResponse{
		DownloadUrl: url,
		ExpiresAt:   time.Now().Add(dc.client.config.PresignExpire).Format(time.RFC3339),
		FileInfo: struct {
			Id      uint   `json:"id"`
			Name    string `json:"name"`
			Size    int64  `json:"size"`
			Hotness int64  `json:"hotness"`
		}{
			Id:      file.ID,
			Name:    file.Name,
			Size:    file.Size,
			Hotness: file.Hotness + 1,
		},
	}, nil
}

func (dc *DownloadClient) GeneratePreviewURL(uid uint, req *dto.DownloadFilesReq) (*dto.PreviewResponse, error) {
	// 可预览的文件类型
	previewableTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"application/pdf": true,
		"text/plain":      true,
	}

	// 获取文件信息
	var file entity.File
	if err := dc.client.db.First(&file, req.FileId).Error; err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	// 验证文件权限
	if file.Uid != uid {
		return nil, fmt.Errorf("无权限访问此文件")
	}

	// 检查文件类型是否支持预览
	if !previewableTypes[file.MineType] {
		return nil, fmt.Errorf("文件类型不支持预览")
	}

	// 生成完整文件路径
	objectKey, err := dc.getFullFilePath(req.FileId)
	if err != nil {
		return nil, fmt.Errorf("获取文件路径失败: %w", err)
	}

	// 生成预览预签名URL
	url, err := dc.generatePreviewPresignedURL(objectKey, file.MineType)
	if err != nil {
		return nil, fmt.Errorf("生成预览链接失败: %w", err)
	}

	return &dto.PreviewResponse{
		PreviewUrl: url,
		ExpiresAt:  time.Now().Add(dc.client.config.PresignExpire).Format(time.RFC3339),
		FileInfo: struct {
			Id          uint   `json:"id"`
			Name        string `json:"name"`
			Size        int64  `json:"size"`
			ContentType string `json:"contentType"`
		}{
			Id:          file.ID,
			Name:        file.Name,
			Size:        file.Size,
			ContentType: file.MineType,
		},
	}, nil
}

func (dc *DownloadClient) generatePresignedURL(objectKey string) (string, error) {
	req, _ := dc.client.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(dc.client.config.BucketName),
		Key:    aws.String(objectKey),
	})
	return req.Presign(dc.client.config.PresignExpire)
}

func (dc *DownloadClient) generatePreviewPresignedURL(objectKey, contentType string) (string, error) {
	req, _ := dc.client.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String(dc.client.config.BucketName),
		Key:                        aws.String(objectKey),
		ResponseContentType:        aws.String(contentType),
		ResponseContentDisposition: aws.String("inline"),
	})
	return req.Presign(dc.client.config.PresignExpire)
}

func (dc *DownloadClient) getFullFilePath(fileId uint) (string, error) {
	var file entity.File
	if err := dc.client.db.First(&file, fileId).Error; err != nil {
		return "", err
	}

	var pathSegments []string
	pathSegments = append(pathSegments, file.Name)

	// 递归查找父文件夹
	currentFolderId := file.FolderId
	for currentFolderId != 0 {
		var folder entity.Folder
		if err := dc.client.db.First(&folder, currentFolderId).Error; err != nil {
			return "", err
		}
		pathSegments = append([]string{folder.Name}, pathSegments...)
		currentFolderId = folder.ParentFolderId
	}

	return strings.Join(pathSegments, "/"), nil
}
