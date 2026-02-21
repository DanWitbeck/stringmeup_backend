// internal/uploads/service.go
package uploads

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/example/threadcraft-backend/internal/config"
	"github.com/google/uuid"
)

type Service struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
	publicURL string
}

func NewService(cfg *config.Config) *Service {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID),
			}, nil
		},
	)

	awsCfg := aws.Config{
		Region:                      "auto",
		Credentials:                 credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, ""),
		EndpointResolverWithOptions: r2Resolver,
	}

	client := s3.NewFromConfig(awsCfg)
	return &Service{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    cfg.R2BucketName,
		publicURL: cfg.R2PublicURL,
	}
}

type PresignResult struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

func (s *Service) Presign(ctx context.Context, userID, contentType string) (*PresignResult, error) {
	key := fmt.Sprintf("users/%s/images/%s", userID, uuid.New().String())

	req, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("presign: %w", err)
	}

	return &PresignResult{
		URL: req.URL,
		Key: fmt.Sprintf("%s/%s", s.publicURL, key),
	}, nil
}
