package client

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"gorm.io/gorm"
	"time"
)

type SealOSConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	BucketName      string
	PresignExpire   time.Duration
}

type SealOSClient struct {
	config     *SealOSConfig
	db         *gorm.DB
	s3Session  *session.Session
	s3Client   *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

func NewSealOSClient(config *SealOSConfig, db *gorm.DB) (*SealOSClient, error) {
	awsConfig := &aws.Config{
		Endpoint:         aws.String(config.Endpoint),
		Region:           aws.String(config.Region),
		Credentials:      credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	return &SealOSClient{
		config:     config,
		db:         db,
		s3Session:  sess,
		s3Client:   s3.New(sess),
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
	}, nil
}

func (c *SealOSClient) GetS3Client() *s3.S3 {
	return c.s3Client
}

func (c *SealOSClient) GetUploader() *s3manager.Uploader {
	return c.uploader
}

func (c *SealOSClient) GetDownloader() *s3manager.Downloader {
	return c.downloader
}

func (c *SealOSClient) GetDB() *gorm.DB {
	return c.db
}

func (c *SealOSClient) GetConfig() *SealOSConfig {
	return c.config
}
