package r2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

type PresignedResponse struct {
	UploadURL string            `json:"upload_url"`
	R2Key     string            `json:"r2_key"`
	ExpiresAt time.Time         `json:"expires_at"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
}

type Client interface {
	GeneratePresignedPutURL(ctx context.Context, key, mimeType string, fileSize int64) (*PresignedResponse, error)
}

type client struct {
	s3Client   *s3.Client
	presignCl  *s3.PresignClient
	bucketName string
}

func NewClient(cfg Config) (Client, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID))
	})
	
	presignCl := s3.NewPresignClient(s3Client)

	return &client{
		s3Client:   s3Client,
		presignCl:  presignCl,
		bucketName: cfg.BucketName,
	}, nil
}

func (c *client) GeneratePresignedPutURL(ctx context.Context, key, mimeType string, fileSize int64) (*PresignedResponse, error) {
	expiration := 5 * time.Minute

	req, err := c.presignCl.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucketName),
		Key:           aws.String(key),
		ContentType:   aws.String(mimeType),
		ContentLength: aws.Int64(fileSize),
	}, func(po *s3.PresignOptions) {
		po.Expires = expiration
	})

	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type":   mimeType,
		"Content-Length": fmt.Sprintf("%d", fileSize),
	}

	return &PresignedResponse{
		UploadURL: req.URL,
		R2Key:     key,
		ExpiresAt: time.Now().Add(expiration),
		Method:    "PUT",
		Headers:   headers,
	}, nil
}
