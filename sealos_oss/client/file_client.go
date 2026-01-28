package client

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"mime/multipart"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/trancecho/ragnarok/sealos_oss/dto"
	"github.com/trancecho/ragnarok/sealos_oss/entity"
)

type FileClient struct {
	client *SealOSClient
}

func NewFileClient(client *SealOSClient) *FileClient {
	return &FileClient{client: client}
}

func (fc *FileClient) AddFile(uid uint, req *dto.AddFileReq, fileHeader *multipart.FileHeader) (*dto.FileResponse, error) {
	// 参数校验
	if req.Name == "" {
		return nil, fmt.Errorf("文件名不能为空")
	}

	// 校验文件名格式
	matched, err := regexp.MatchString(`^\d{8}_.+`, req.Name)
	if err != nil || !matched {
		return nil, fmt.Errorf("文件名格式不正确，请使用格式：YYYYMMDD_描述.扩展名")
	}

	// 校验日期部分
	if len(req.Name) >= 8 {
		dateStr := req.Name[:8]
		if _, err := time.Parse("20060102", dateStr); err != nil {
			return nil, fmt.Errorf("文件名中的日期部分无效，请使用有效的YYYYMMDD格式日期")
		}
	}

	// 检查是否已存在同名文件
	var existingFile entity.File
	affected := fc.client.db.Where("uid = ? AND folder_id = ? AND name = ?", uid, req.FolderId, req.Name).
		First(&existingFile).RowsAffected
	if affected > 0 {
		return nil, fmt.Errorf("该文件名已存在，请勿重复上传")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 创建文件实体
	newFile := entity.File{
		Uid:       uid,
		FolderId:  req.FolderId,
		Name:      req.Name,
		MineType:  fileHeader.Header.Get("Content-Type"),
		Size:      fileHeader.Size,
		Hotness:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 生成对象存储路径
	objectKey := fc.getObjectKey(req.FolderId, req.Name)

	// 开始事务
	tx := fc.client.db.Begin()

	// 创建数据库记录
	if err := tx.Create(&newFile).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("数据库操作失败: %w", err)
	}

	// 上传文件到对象存储
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	_, err = fc.client.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(fc.client.config.BucketName),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(newFile.MineType),
	})

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("文件上传失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return &dto.FileResponse{
		Id:   newFile.ID,
		Name: newFile.Name,
		Size: newFile.Size,
		Path: objectKey,
	}, nil
}

func (fc *FileClient) UpdateFileName(uid uint, req *dto.UpdateFileReq) (*dto.FileResponse, error) {
	if req.Id == 0 {
		return nil, fmt.Errorf("文件ID不能为空")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("文件名不能为空")
	}

	// 检查文件是否存在并属于当前用户
	var file entity.File
	if err := fc.client.db.Where("id = ? AND uid = ?", req.Id, uid).First(&file).Error; err != nil {
		return nil, fmt.Errorf("文件不存在或无权访问")
	}

	// 检查新文件名是否已存在
	var existingFile entity.File
	affected := fc.client.db.Where("uid = ? AND folder_id = ? AND name = ? AND id != ?",
		uid, file.FolderId, req.Name, req.Id).First(&existingFile).RowsAffected
	if affected > 0 {
		return nil, fmt.Errorf("该文件名已存在，请使用其他名称")
	}

	// 更新文件名
	file.Name = req.Name
	file.UpdatedAt = time.Now()

	if err := fc.client.db.Save(&file).Error; err != nil {
		return nil, fmt.Errorf("更新文件名失败: %w", err)
	}

	return &dto.FileResponse{
		Id:   file.ID,
		Name: file.Name,
		Size: file.Size,
	}, nil
}

func (fc *FileClient) DeleteFile(uid uint, req *dto.DeleteFileReq) error {
	if req.Id == 0 {
		return fmt.Errorf("文件ID不能为空")
	}

	// 检查文件是否存在并属于当前用户
	var file entity.File
	if err := fc.client.db.Where("id = ? AND uid = ?", req.Id, uid).First(&file).Error; err != nil {
		return fmt.Errorf("文件不存在或无权访问")
	}

	// 开始事务
	tx := fc.client.db.Begin()

	// 从数据库删除文件记录
	if err := tx.Delete(&file).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("数据库删除失败: %w", err)
	}

	// 从对象存储删除文件
	objectKey := fc.getObjectKey(file.FolderId, file.Name)
	_, err := fc.client.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(fc.client.config.BucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("对象存储删除失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

func (fc *FileClient) GetFilesFromFolder(uid uint, req *dto.GetFilesReq) (map[string]interface{}, error) {
	// 设置默认分页值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 验证排序参数
	if req.SortBy != "" && req.SortBy != "new" && req.SortBy != "hot" {
		return nil, fmt.Errorf("sortby参数必须是'new'或'hot'")
	}

	// 获取文件总数
	var totalCount int64
	if err := fc.client.db.Model(&entity.File{}).
		Where("folder_id = ? and uid = ?", req.FolderId, uid).
		Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("获取文件总数失败: %w", err)
	}

	// 计算总页数
	totalPages := totalCount / int64(req.PageSize)
	if totalCount%int64(req.PageSize) > 0 {
		totalPages++
	}

	offset := (req.Page - 1) * req.PageSize

	responseData := map[string]interface{}{
		"pagination": map[string]interface{}{
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total":       totalCount,
			"total_pages": totalPages,
		},
		"current_folder_id": req.FolderId,
	}

	var filesNew []entity.File
	var filesHot []entity.File

	switch req.SortBy {
	case "new":
		if err := fc.client.db.Where("folder_id = ? and uid = ?", req.FolderId, uid).
			Order("SUBSTRING(name, 1, 8) DESC").
			Offset(offset).
			Limit(req.PageSize).
			Find(&filesNew).Error; err != nil {
			return nil, fmt.Errorf("获取文件失败: %w", err)
		}
		responseData["sortby"] = map[string]interface{}{"new": filesNew}

	case "hot":
		if err := fc.client.db.Where("folder_id = ? and uid = ?", req.FolderId, uid).
			Order("hotness DESC").
			Offset(offset).
			Limit(req.PageSize).
			Find(&filesHot).Error; err != nil {
			return nil, fmt.Errorf("获取文件失败: %w", err)
		}
		responseData["sortby"] = map[string]interface{}{"hot": filesHot}

	default:
		if err := fc.client.db.Where("folder_id = ? and uid = ?", req.FolderId, uid).
			Order("created_at DESC").
			Offset(offset).
			Limit(req.PageSize).
			Find(&filesNew).Error; err != nil {
			return nil, fmt.Errorf("获取文件失败: %w", err)
		}

		if err := fc.client.db.Where("folder_id = ? and uid = ?", req.FolderId, uid).
			Order("hotness DESC").
			Offset(offset).
			Limit(req.PageSize).
			Find(&filesHot).Error; err != nil {
			return nil, fmt.Errorf("获取文件失败: %w", err)
		}

		responseData["sortby"] = map[string]interface{}{
			"new": filesNew,
			"hot": filesHot,
		}
	}

	return responseData, nil
}

func (fc *FileClient) getObjectKey(folderId uint, fileName string) string {
	// 简化的路径生成逻辑，实际应根据需要实现完整的路径生成
	return fmt.Sprintf("%d/%s", folderId, fileName)
}
