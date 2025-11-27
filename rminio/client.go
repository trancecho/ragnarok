package rminio

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient MinIO客户端结构体
type MinioClient struct {
	client *minio.Client
}

// NewMinioClient 初始化MinIO客户端
func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*MinioClient, error) {
	// 初始化minio client对象
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %v", err)
	}

	return &MinioClient{client: client}, nil
}

// BucketExists 检查存储桶是否存在
func (m *MinioClient) BucketExists(bucketName string) (bool, error) {
	exists, err := m.client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return false, fmt.Errorf("failed to check if bucket exists: %v", err)
	}
	return exists, nil
}

// MakeBucket 创建存储桶
func (m *MinioClient) MakeBucket(bucketName string, location string) error {
	err := m.client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %v", err)
	}
	return nil
}

// UploadFile 上传文件到MinIO
func (m *MinioClient) UploadFile(bucketName, objectName, filePath string, contentType string) error {
	// 上传文件
	uploadInfo, err := m.client.FPutObject(context.Background(), bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	log.Printf("Successfully uploaded %s of size %d\n", objectName, uploadInfo.Size)
	return nil
}

// UploadFromStream 从流上传文件到MinIO
func (m *MinioClient) UploadFromStream(bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := m.client.PutObject(context.Background(), bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload from stream: %v", err)
	}
	log.Printf("Successfully uploaded %s from stream\n", objectName)
	return nil
}

// DownloadFile 下载文件
func (m *MinioClient) DownloadFile(bucketName, objectName, filePath string) error {
	err := m.client.FGetObject(context.Background(), bucketName, objectName, filePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	log.Printf("Successfully downloaded %s to %s\n", objectName, filePath)
	return nil
}

// DownloadAsStream 下载文件为流
func (m *MinioClient) DownloadAsStream(bucketName, objectName string) (*minio.Object, error) {
	obj, err := m.client.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download as stream: %v", err)
	}
	return obj, nil
}

// GetFileContent 获取文件内容为字节数组
func (m *MinioClient) GetFileContent(bucketName, objectName string) ([]byte, error) {
	obj, err := m.client.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %v", err)
	}
	defer obj.Close()

	// 读取所有内容
	content, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object content: %v", err)
	}

	return content, nil
}

// RemoveFile 删除文件
func (m *MinioClient) RemoveFile(bucketName, objectName string) error {
	err := m.client.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove file: %v", err)
	}
	log.Printf("Successfully removed %s\n", objectName)
	return nil
}

// RemoveFiles 批量删除文件
func (m *MinioClient) RemoveFiles(bucketName string, objectNames []string) error {
	objectsCh := make(chan minio.ObjectInfo)

	// 发送要删除的对象名称
	go func() {
		defer close(objectsCh)
		for _, objectName := range objectNames {
			objectsCh <- minio.ObjectInfo{Key: objectName}
		}
	}()

	// 调用RemoveObjects删除多个对象
	for rErr := range m.client.RemoveObjects(context.Background(), bucketName, objectsCh, minio.RemoveObjectsOptions{}) {
		return fmt.Errorf("error detected during deletion: %v", rErr.Err)
	}

	log.Printf("Successfully removed %d files\n", len(objectNames))
	return nil
}

// PresignedGetObject 生成预签名URL用于下载
func (m *MinioClient) PresignedGetObject(bucketName, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(context.Background(), bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}
	return url.String(), nil
}

// PresignedPutObject 生成预签名URL用于上传
func (m *MinioClient) PresignedPutObject(bucketName, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedPutObject(context.Background(), bucketName, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}
	return url.String(), nil
}

// ListObjects 列出存储桶中的对象
func (m *MinioClient) ListObjects(bucketName, prefix string, recursive bool) <-chan minio.ObjectInfo {
	return m.client.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})
}

// StatObject 获取对象信息
func (m *MinioClient) StatObject(bucketName, objectName string) (minio.ObjectInfo, error) {
	info, err := m.client.StatObject(context.Background(), bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to stat object: %v", err)
	}
	return info, nil
}

func (m *MinioClient) FileExists(bucketName, objectName string) (bool, error) {
	_, err := m.client.StatObject(context.Background(), bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil // 文件不存在
		}
		return false, fmt.Errorf("failed to check if file exists: %v", err)
	}
	return true, nil // 文件存在
}
