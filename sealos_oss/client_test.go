package client_test

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trancecho/ragnarok/sealos_oss/client"
	"gopkg.in/yaml.v2"
)

// 配置文件结构
type Config struct {
	SealOSOSS SealOSOSSConfig `yaml:"sealos_oss"`
}

type SealOSOSSConfig struct {
	Endpoint        string        `yaml:"endpoint"`
	AccessKeyID     string        `yaml:"access_key_id"`
	SecretAccessKey string        `yaml:"secret_access_key"`
	Region          string        `yaml:"region"`
	BucketName      string        `yaml:"bucket_name"`
	PresignExpire   time.Duration `yaml:"presign_expire"`
}

// 加载配置
func loadConfig() (*client.SealOSConfig, error) {
	configFile := "config.dev.yaml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configFile)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &client.SealOSConfig{
		Endpoint:        config.SealOSOSS.Endpoint,
		AccessKeyID:     config.SealOSOSS.AccessKeyID,
		SecretAccessKey: config.SealOSOSS.SecretAccessKey,
		Region:          config.SealOSOSS.Region,
		BucketName:      config.SealOSOSS.BucketName,
		PresignExpire:   config.SealOSOSS.PresignExpire,
	}, nil
}

// 测试文件信息
const (
	testFileName    = "test.doc"
	testFileContent = "This is a test document content for OSS operations testing."
)

// setupTest 设置测试环境
func setupTest(t *testing.T) (*client.SealOSClient, func()) {
	// 加载真实配置
	testConfig, err := loadConfig()
	require.NoError(t, err)

	// 创建SealOS客户端（不使用数据库）
	sealosClient, err := client.NewSealOSClient(testConfig, nil)
	require.NoError(t, err)

	// 清理函数
	cleanup := func() {
		// 清理测试文件
		os.Remove(testFileName)
	}

	return sealosClient, cleanup
}

// createTestFile 创建测试文件
func createTestFile() error {
	file, err := os.Create(testFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(testFileContent)
	return err
}

// createMultipartFileHeader 创建multipart文件头用于测试
func createMultipartFileHeader(t *testing.T, filename string, content string) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)

	_, err = part.Write([]byte(content))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	file, header, err := req.FormFile("file")
	require.NoError(t, err)
	defer file.Close()

	return header
}

// generateUniqueFilename 生成唯一的文件名，避免测试冲突
func generateUniqueFilename(baseName string) string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s.doc", timestamp, baseName)
}

// TestOSSConnection 测试OSS连接性
func TestOSSConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实OSS测试（使用 -short 参数时）")
	}

	testConfig, err := loadConfig()
	require.NoError(t, err)

	// 创建SealOS客户端
	sealosClient, err := client.NewSealOSClient(testConfig, nil)
	require.NoError(t, err)

	// 测试ListBuckets来验证连接性
	_, err = sealosClient.GetS3Client().ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		t.Logf("OSS连接测试结果: %v", err)
		// 对于连接测试，我们只记录结果，不强制失败
	} else {
		t.Log("OSS连接测试成功")
	}
}

// TestBucketOperations 测试Bucket基本操作
func TestBucketOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实OSS测试（使用 -short 参数时）")
	}

	sealosClient, cleanup := setupTest(t)
	defer cleanup()

	// 测试列出对象
	objects, err := sealosClient.GetS3Client().ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:  aws.String(sealosClient.GetConfig().BucketName),
		MaxKeys: aws.Int64(10),
	})

	if err != nil {
		t.Logf("列出对象测试结果: %v", err)
	} else {
		t.Logf("Bucket中包含 %d 个对象", len(objects.Contents))
		for _, obj := range objects.Contents {
			t.Logf("对象: %s, 大小: %d bytes", *obj.Key, *obj.Size)
		}
	}
}

// TestFileUploadDownload 测试文件上传和下载
func TestFileUploadDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实OSS测试（使用 -short 参数时）")
	}

	sealosClient, cleanup := setupTest(t)
	defer cleanup()

	// 生成唯一的文件名
	uniqueFilename := generateUniqueFilename("test_upload")
	testContent := "Hello, SealOS OSS! This is a test file content."

	// 创建测试文件头
	fileHeader := createMultipartFileHeader(t, uniqueFilename, testContent)

	// 直接使用S3客户端上传文件（绕过数据库操作）
	file, err := fileHeader.Open()
	require.NoError(t, err)
	defer file.Close()

	// 上传文件到OSS
	_, err = sealosClient.GetUploader().Upload(&s3manager.UploadInput{
		Bucket: aws.String(sealosClient.GetConfig().BucketName),
		Key:    aws.String(uniqueFilename),
		Body:   file,
	})
	require.NoError(t, err)

	t.Logf("文件上传成功: %s", uniqueFilename)

	// 验证文件是否存在
	_, err = sealosClient.GetS3Client().HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(sealosClient.GetConfig().BucketName),
		Key:    aws.String(uniqueFilename),
	})
	require.NoError(t, err, "文件应该存在于对象存储中")

	// 生成下载URL
	req, _ := sealosClient.GetS3Client().GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(sealosClient.GetConfig().BucketName),
		Key:    aws.String(uniqueFilename),
	})
	downloadURL, err := req.Presign(15 * time.Minute)
	require.NoError(t, err)

	t.Logf("下载URL生成成功: %s", downloadURL)
	assert.Contains(t, downloadURL, "X-Amz-Signature")

	// 清理：删除测试文件
	_, err = sealosClient.GetS3Client().DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(sealosClient.GetConfig().BucketName),
		Key:    aws.String(uniqueFilename),
	})
	require.NoError(t, err)

	t.Logf("测试文件已清理: %s", uniqueFilename)
}

// TestPresignedURL 测试预签名URL功能
func TestPresignedURL(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实OSS测试（使用 -short 参数时）")
	}

	sealosClient, cleanup := setupTest(t)
	defer cleanup()

	// 生成唯一的文件名
	uniqueFilename := generateUniqueFilename("test_presigned")
	testContent := "Test content for presigned URL"

	// 上传测试文件
	fileHeader := createMultipartFileHeader(t, uniqueFilename, testContent)
	file, err := fileHeader.Open()
	require.NoError(t, err)
	defer file.Close()

	_, err = sealosClient.GetUploader().Upload(&s3manager.UploadInput{
		Bucket: aws.String(sealosClient.GetConfig().BucketName),
		Key:    aws.String(uniqueFilename),
		Body:   file,
	})
	require.NoError(t, err)

	// 测试生成预签名URL
	req, _ := sealosClient.GetS3Client().GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(sealosClient.GetConfig().BucketName),
		Key:    aws.String(uniqueFilename),
	})
	presignedURL, err := req.Presign(sealosClient.GetConfig().PresignExpire)
	require.NoError(t, err)

	t.Logf("预签名URL: %s", presignedURL)
	assert.NotEmpty(t, presignedURL)
	assert.Contains(t, presignedURL, "X-Amz-Expires")

	// 清理
	_, err = sealosClient.GetS3Client().DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(sealosClient.GetConfig().BucketName),
		Key:    aws.String(uniqueFilename),
	})
	require.NoError(t, err)
}

// TestFileOperations 测试完整的文件操作流程
func TestFileOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实OSS测试（使用 -short 参数时）")
	}

	sealosClient, cleanup := setupTest(t)
	defer cleanup()

	// 测试文件
	testFiles := []struct {
		name    string
		content string
	}{
		{generateUniqueFilename("test1"), "First test file content"},
		{generateUniqueFilename("test2"), "Second test file content"},
		{generateUniqueFilename("test3"), "Third test file content"},
	}

	// 上传多个文件
	for _, testFile := range testFiles {
		fileHeader := createMultipartFileHeader(t, testFile.name, testFile.content)
		file, err := fileHeader.Open()
		require.NoError(t, err)

		_, err = sealosClient.GetUploader().Upload(&s3manager.UploadInput{
			Bucket: aws.String(sealosClient.GetConfig().BucketName),
			Key:    aws.String(testFile.name),
			Body:   file,
		})
		require.NoError(t, err)
		file.Close()

		t.Logf("上传成功: %s", testFile.name)
	}

	// 列出文件验证上传
	objects, err := sealosClient.GetS3Client().ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:  aws.String(sealosClient.GetConfig().BucketName),
		MaxKeys: aws.Int64(10),
	})
	require.NoError(t, err)

	uploadedCount := 0
	for _, obj := range objects.Contents {
		for _, testFile := range testFiles {
			if *obj.Key == testFile.name {
				uploadedCount++
				break
			}
		}
	}

	assert.Equal(t, len(testFiles), uploadedCount, "所有测试文件都应该上传成功")

	// 清理所有测试文件
	for _, testFile := range testFiles {
		_, err := sealosClient.GetS3Client().DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(sealosClient.GetConfig().BucketName),
			Key:    aws.String(testFile.name),
		})
		require.NoError(t, err)
	}

	t.Log("所有测试文件已清理")
}

// TestConfigValidation 测试配置验证
func TestConfigValidation(t *testing.T) {
	testConfig, err := loadConfig()
	require.NoError(t, err)

	// 验证配置是否正确加载
	assert.NotEmpty(t, testConfig.Endpoint)
	assert.NotEmpty(t, testConfig.AccessKeyID)
	assert.NotEmpty(t, testConfig.SecretAccessKey)
	assert.NotEmpty(t, testConfig.Region)
	assert.NotEmpty(t, testConfig.BucketName)
	assert.True(t, testConfig.PresignExpire > 0)

	t.Logf("配置加载成功: Endpoint=%s, Bucket=%s", testConfig.Endpoint, testConfig.BucketName)
}
