package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3ClientAPI interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type Service struct {
	client S3ClientAPI
}

// Load the Shared AWS Configuration (~/.aws/config)
func NewAWSConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, err
	}

	return cfg, nil
}

func NewClient(cfg aws.Config) *s3.Client {
	return s3.NewFromConfig(cfg)
}

func NewService(client S3ClientAPI) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) GetClient() *s3.Client {
	return s.client.(*s3.Client)
}

func (s *Service) ListObjects(ctx context.Context, bucket *string, prefix *string, startAfter *string) (*s3.ListObjectsV2Output, error) {
	result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:     bucket,
		Prefix:     prefix,
		StartAfter: startAfter,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) GetObject(ctx context.Context, bucket *string, key *string) (*s3.GetObjectOutput, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) PutObject(ctx context.Context, bucket *string, key *string, body io.Reader) (*s3.PutObjectOutput, error) {
	result, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: bucket,
		Key:    key,
		Body:   body,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) DeleteObject(ctx context.Context, bucket *string, key *string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		return err
	}

	return nil
}
