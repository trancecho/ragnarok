package rminio

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
)

// UploadBlobFromDB 从数据库读取Blob并上传到MinIO
func (m *MinioClient) UploadBlobFromDB(db *sql.DB, bucketName, objectName string, query string, args ...interface{}) error {
	// 从数据库查询Blob数据
	var data []byte
	var mimeType string

	err := db.QueryRow(query, args...).Scan(&data, &mimeType)
	if err != nil {
		return fmt.Errorf("failed to query blob from database: %v", err)
	}

	// 创建bytes.Reader作为数据流
	reader := bytes.NewReader(data)
	objectSize := int64(len(data))

	// 上传到MinIO
	_, err = m.client.PutObject(context.Background(), bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload blob to minio: %v", err)
	}

	log.Printf("Successfully uploaded blob as %s to bucket %s\n", objectName, bucketName)
	return nil
}

// UploadBlobDirect 直接上传Blob数据到MinIO
func (m *MinioClient) UploadBlobDirect(bucketName, objectName string, blobData []byte, mimeType string) error {
	reader := bytes.NewReader(blobData)
	objectSize := int64(len(blobData))

	_, err := m.client.PutObject(context.Background(), bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload blob to minio: %v", err)
	}

	log.Printf("Successfully uploaded blob as %s to bucket %s\n", objectName, bucketName)
	return nil
}
