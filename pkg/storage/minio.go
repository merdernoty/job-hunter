package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/merdernoty/job-hunter/config"
	"github.com/merdernoty/job-hunter/pkg/logger"
)

type MinIOClient struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	logger     logger.Logger
}

func NewMinIOClient(cfg *config.Config, logger logger.Logger) (*MinIOClient, error) {
	logger.Infof("Initializing MinIO client with endpoint: %s", cfg.MiniO.Endpoint)
	
	client, err := minio.New(cfg.MiniO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MiniO.AccessKeyID, cfg.MiniO.SecretAccessKey, ""),
		Secure: cfg.MiniO.UseSSL,
		Region: cfg.MiniO.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	minioClient := &MinIOClient{
		client:     client,
		bucketName: cfg.MiniO.BucketName,
		endpoint:   cfg.MiniO.Endpoint,
		logger:     logger,
	}

	if err := minioClient.testConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO: %w", err)
	}

	if err := minioClient.ensureBucketExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	if err := minioClient.setBucketPolicy(); err != nil {
		logger.Warnf("Failed to set bucket policy: %v", err)
	}

	logger.Infof("MinIO client initialized successfully with bucket: %s", cfg.MiniO.BucketName)
	return minioClient, nil
}

func (m *MinIOClient) testConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to MinIO server: %w", err)
	}

	m.logger.Info("Successfully connected to MinIO server")
	return nil
}

func (m *MinIOClient) ensureBucketExists() error {
	ctx := context.Background()
	
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{
			Region: "us-east-1",
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		m.logger.Infof("Created MinIO bucket: %s", m.bucketName)
	} else {
		m.logger.Infof("MinIO bucket already exists: %s", m.bucketName)
	}

	return nil
}

func (m *MinIOClient) setBucketPolicy() error {
	ctx := context.Background()
	
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": "*",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::%s/avatars/*"
			}
		]
	}`, m.bucketName)

	err := m.client.SetBucketPolicy(ctx, m.bucketName, policy)
	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	m.logger.Infof("Set public read policy for bucket: %s", m.bucketName)
	return nil
}

func (m *MinIOClient) GetClient() *minio.Client {
	return m.client
}

func (m *MinIOClient) GetBucketName() string {
	return m.bucketName
}

func (m *MinIOClient) GetEndpoint() string {
	return m.endpoint
}