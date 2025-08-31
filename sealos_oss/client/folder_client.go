package client

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/trancecho/ragnarok/sealos_oss/dto"
	"github.com/trancecho/ragnarok/sealos_oss/entity"
)

type FolderClient struct {
	client *SealOSClient
}

func NewFolderClient(client *SealOSClient) *FolderClient {
	return &FolderClient{client: client}
}

func (fc *FolderClient) CreateFolder(uid uint, req *dto.AddFolderReq) (*dto.FolderOperationResponse, error) {
	// 检查同名文件夹是否已存在
	var count int64
	if err := fc.client.db.Model(&entity.Folder{}).
		Where("uid = ? AND parent_folder_id = ? AND name = ?", uid, req.ParentFolderId, req.Name).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("检查文件夹名称失败: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("该名称的文件夹已存在")
	}

	// 创建文件夹实体
	folder := entity.Folder{
		Uid:            uid,
		Name:           req.Name,
		ParentFolderId: req.ParentFolderId,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 获取文件夹路径
	folderPath := fc.getFolderPath(&folder)
	folderKey := folderPath + "/"

	// 开始事务
	tx := fc.client.db.Begin()

	// 保存到数据库
	if err := tx.Create(&folder).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建文件夹记录失败: %w", err)
	}

	// 在对象存储中创建文件夹标记
	if err := fc.createFolderInObjectStorage(folderKey); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("在对象存储中创建文件夹失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return &dto.FolderOperationResponse{
		Success:   true,
		Message:   "文件夹创建成功",
		FolderId:  folder.ID,
		FolderKey: folderKey,
	}, nil
}

func (fc *FolderClient) GetSonFolders(uid uint, req *dto.GetSonFoldersReq) (map[string]interface{}, error) {
	var folders []entity.Folder
	var parentFolder entity.Folder

	if req.Id != 0 {
		affected := fc.client.db.Where("id = ? and uid = ?", req.Id, uid).First(&parentFolder).RowsAffected
		if affected == 0 {
			return nil, fmt.Errorf("父文件夹不存在")
		}
	}

	affected := fc.client.db.Where("parent_folder_id = ? and uid = ?", req.Id, uid).Find(&folders).RowsAffected
	if affected == 0 {
		folders = []entity.Folder{} // 返回空数组而不是错误
	}

	parentFolderId := parentFolder.ParentFolderId

	return map[string]interface{}{
		"folders":             folders,
		"current_folder_id":   req.Id,
		"parent_folder_id":    parentFolderId,
		"current_folder_name": parentFolder.Name,
	}, nil
}

func (fc *FolderClient) UpdateFolderName(uid uint, req *dto.UpdateFolderReq) (*dto.FolderOperationResponse, error) {
	// 获取要修改的文件夹信息
	folder, err := fc.getFolderById(req.Id)
	if err != nil {
		return nil, fmt.Errorf("获取文件夹信息失败: %w", err)
	}

	// 验证权限
	if folder.Uid != uid {
		return nil, fmt.Errorf("无权限修改此文件夹")
	}

	// 检查新名称是否与现有文件夹冲突
	var count int64
	if err := fc.client.db.Model(&entity.Folder{}).
		Where("uid = ? AND parent_folder_id = ? AND name = ? AND id != ?",
			uid, folder.ParentFolderId, req.Name, req.Id).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("检查文件夹名称失败: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("该名称的文件夹已存在")
	}

	// 获取旧路径和新路径
	oldPath := fc.getFolderPath(folder)
	oldKey := oldPath + "/"

	originalName := folder.Name
	folder.Name = req.Name
	newPath := fc.getFolderPath(folder)
	newKey := newPath + "/"
	folder.Name = originalName // 恢复原名

	// 开始事务
	tx := fc.client.db.Begin()

	// 更新数据库中的文件夹名称
	if err := tx.Model(&entity.Folder{}).
		Where("id = ?", req.Id).
		Update("name", req.Name).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新文件夹名称失败: %w", err)
	}

	// 更新对象存储中的文件夹标记
	if err := fc.updateFolderInObjectStorage(oldKey, newKey); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新对象存储中的文件夹失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return &dto.FolderOperationResponse{
		Success:   true,
		Message:   "文件夹名称更新成功",
		FolderId:  folder.ID,
		FolderKey: newKey,
	}, nil
}

func (fc *FolderClient) DeleteFolder(uid uint, req *dto.DeleteFolderReq) error {
	// 获取文件夹信息
	folder, err := fc.getFolderById(req.Id)
	if err != nil {
		return fmt.Errorf("获取文件夹信息失败: %w", err)
	}

	// 验证权限
	if folder.Uid != uid {
		return fmt.Errorf("无权限删除此文件夹")
	}

	// 获取文件夹路径
	folderPath := fc.getFolderPath(folder)
	folderKey := folderPath + "/"

	// 开始事务
	tx := fc.client.db.Begin()

	// 递归删除对象存储中的内容
	if err := fc.deleteFolderContentsFromObjectStorage(folderKey); err != nil {
		tx.Rollback()
		return fmt.Errorf("从对象存储删除文件夹内容失败: %w", err)
	}

	// 删除数据库中的文件夹记录
	if err := tx.Delete(&folder).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("从数据库删除文件夹记录失败: %w", err)
	}

	// 删除所有子文件夹记录
	if err := tx.Where("parent_folder_id = ?", req.Id).Delete(&entity.Folder{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除子文件夹记录失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

func (fc *FolderClient) createFolderInObjectStorage(folderKey string) error {
	_, err := fc.client.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(fc.client.config.BucketName),
		Key:    aws.String(folderKey),
	})
	return err
}

func (fc *FolderClient) updateFolderInObjectStorage(oldKey, newKey string) error {
	// 检查旧文件夹是否存在
	_, err := fc.client.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(fc.client.config.BucketName),
		Key:    aws.String(oldKey),
	})
	if err != nil {
		// 如果旧文件夹不存在，可能不需要更新
		return nil
	}

	// 创建新文件夹标记
	_, err = fc.client.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(fc.client.config.BucketName),
		Key:    aws.String(newKey),
	})
	if err != nil {
		return err
	}

	// 删除旧文件夹标记
	_, err = fc.client.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(fc.client.config.BucketName),
		Key:    aws.String(oldKey),
	})

	return err
}

func (fc *FolderClient) deleteFolderContentsFromObjectStorage(prefix string) error {
	// 列出所有对象
	var objectKeys []*s3.ObjectIdentifier
	err := fc.client.s3Client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(fc.client.config.BucketName),
		Prefix: aws.String(prefix),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			objectKeys = append(objectKeys, &s3.ObjectIdentifier{
				Key: obj.Key,
			})
		}
		return !lastPage
	})
	if err != nil {
		return err
	}

	// 批量删除对象
	if len(objectKeys) > 0 {
		_, err = fc.client.s3Client.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(fc.client.config.BucketName),
			Delete: &s3.Delete{
				Objects: objectKeys,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (fc *FolderClient) getFolderById(id uint) (*entity.Folder, error) {
	var folder entity.Folder
	if err := fc.client.db.First(&folder, id).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}

func (fc *FolderClient) getFolderPath(folder *entity.Folder) string {
	if folder.ParentFolderId == 0 {
		return folder.Name
	}

	parent, err := fc.getFolderById(folder.ParentFolderId)
	if err != nil {
		return folder.Name
	}

	return fc.getFolderPath(parent) + "/" + folder.Name
}
