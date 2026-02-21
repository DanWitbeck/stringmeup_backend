// internal/uploads/service.go
package uploads

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/google/uuid"
	"stringmeup/backend/internal/config"
)

// r2EndpointResolver routes all S3 calls to Cloudflare R2
type r2EndpointResolver struct {
	accountID string
}

func (r *r2EndpointResolver) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (
	smithyendpoints.Endpoint, error) {
	return smithyendpoints.Endpoint{
		URI: *aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r.accountID)),
	}, nil
}

type Service struct {
	presigner *s3.PresignClient
	bucket    string
	publicURL string
}

func NewService(cfg *config.Config) *Service {
	client := s3.New(s3.Options{
		Region: "auto",
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.R2AccessKeyID, cfg.R2SecretAccessKey, ""),
		EndpointResolverV2: &r2EndpointResolver{accountID: cfg.R2AccountID},
	})

	return &Service{
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
