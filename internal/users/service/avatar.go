package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/merdernoty/job-hunter/pkg/logger"
	minioClient "github.com/merdernoty/job-hunter/pkg/storage"
	"github.com/minio/minio-go/v7"
)

type AvatarService struct {
	minioClient *minioClient.MinIOClient
	logger      logger.Logger
}

func NewAvatarService(minioClient *minioClient.MinIOClient, logger logger.Logger) *AvatarService {
	return &AvatarService{
		minioClient: minioClient,
		logger:      logger,
	}
}

type UploadAvatarRequest struct {
	UserID      uuid.UUID
	File        io.Reader
	FileName    string
	FileSize    int64
	ContentType string
}

func (s *AvatarService) UploadAvatar(req UploadAvatarRequest) (string, error) {
	if err := s.validateAvatarRequest(req); err != nil {
		return "", err
	}

	objectName := s.generateAvatarPath(req.UserID, req.FileName)

	if err := s.uploadToMinio(objectName, req.File, req.FileSize, req.ContentType); err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	avatarURL := s.generatePublicURL(objectName)

	s.logger.Infof("Successfully uploaded avatar for user %s: %s", req.UserID, avatarURL)
	return avatarURL, nil
}

func (s *AvatarService) DeleteAvatar(avatarURL string) error {
	objectName := s.extractObjectNameFromURL(avatarURL)
	if objectName == "" {
		return fmt.Errorf("invalid avatar URL")
	}

	ctx := context.Background()
	err := s.minioClient.GetClient().RemoveObject(ctx, s.minioClient.GetBucketName(), objectName, minio.RemoveObjectOptions{})
	if err != nil {
		s.logger.Errorf("Failed to delete avatar %s: %v", objectName, err)
		return fmt.Errorf("failed to delete avatar from storage")
	}

	s.logger.Infof("Successfully deleted avatar: %s", objectName)
	return nil
}

func (s *AvatarService) validateAvatarRequest(req UploadAvatarRequest) error {
	maxSize := int64(2 * 1024 * 1024) // 2MB
	if req.FileSize > maxSize {
		return fmt.Errorf("avatar file too large: maximum size is 2MB")
	}

	if req.FileSize <= 0 {
		return fmt.Errorf("invalid file size")
	}

	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !allowedTypes[strings.ToLower(req.ContentType)] {
		return fmt.Errorf("invalid file type: only images are allowed (jpeg, png, gif, webp)")
	}

	return nil
}

func (s *AvatarService) generateAvatarPath(userID uuid.UUID, originalFileName string) string {
	ext := filepath.Ext(originalFileName)
	if ext == "" {
		ext = ".jpg"
	}
	ext = strings.ToLower(ext)

	timestamp := time.Now().Unix()
	uniqueID := uuid.New().String()

	return fmt.Sprintf("avatars/%s/%d_%s%s", userID.String(), timestamp, uniqueID, ext)
}

func (s *AvatarService) uploadToMinio(objectName string, file io.Reader, fileSize int64, contentType string) error {
	ctx := context.Background()

	opts := minio.PutObjectOptions{
		ContentType:  contentType,
		CacheControl: "max-age=31536000",
	}

	_, err := s.minioClient.GetClient().PutObject(
		ctx,
		s.minioClient.GetBucketName(),
		objectName,
		file,
		fileSize,
		opts,
	)

	if err != nil {
		s.logger.Errorf("Failed to upload to MinIO: %v", err)
		return err
	}

	return nil
}

func (s *AvatarService) generatePublicURL(objectName string) string {
	endpoint := s.minioClient.GetEndpoint()
	bucketName := s.minioClient.GetBucketName()

	if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	}

	return fmt.Sprintf("http://%s/%s/%s", endpoint, bucketName, objectName)
}

func (s *AvatarService) extractObjectNameFromURL(avatarURL string) string {
	parts := strings.Split(avatarURL, "/")
	if len(parts) < 3 {
		return ""
	}

	bucketName := s.minioClient.GetBucketName()
	foundBucket := false
	var objectParts []string

	for _, part := range parts {
		if foundBucket {
			objectParts = append(objectParts, part)
		} else if part == bucketName {
			foundBucket = true
		}
	}

	if len(objectParts) == 0 {
		return ""
	}

	return strings.Join(objectParts, "/")
}
