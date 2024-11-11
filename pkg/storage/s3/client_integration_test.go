package s3_test

import (
	"context"
	"go_micro_service_api/pkg/storage/s3"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#creating-the-credentials-file
	cfg, err := s3.NewAWSConfig(context.Background())
	assert.Nil(t, err)

	// https://stackoverflow.com/questions/73987705/aws-s3client-does-not-load-credentials-properly
	cred, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Skip("AWS credentials not found")
	}
	if cred.AccessKeyID != "" && cred.SecretAccessKey != "" {
		t.Skip("AWS credentials not found")
	}

	client := s3.NewClient(cfg)
	assert.NotNil(t, client)

	service := s3.NewService(client)

	t.Run("ListObjects", func(t *testing.T) {
		bucket := "kyc4dev"
		prefix := "test/"
		_, err := service.ListObjects(context.Background(), &bucket, &prefix, nil)
		assert.Nil(t, err)
	})

	t.Run("GetObject", func(t *testing.T) {
		bucket := "go_micro_service_api"
		key := "test/test.txt"
		_, err := service.GetObject(context.Background(), &bucket, &key)
		assert.Nil(t, err)
	})

	t.Run("PutObject", func(t *testing.T) {
		bucket := "go_micro_service_api"
		key := "test/test.txt"
		body := "Hello, World!"
		result, err := service.PutObject(context.Background(), &bucket, &key, strings.NewReader(body))
		assert.Nil(t, err)

		assert.NotNil(t, result.VersionId)
	})

	t.Run("DeleteObject", func(t *testing.T) {
		bucket := "go_micro_service_api"
		key := "test/test.txt"
		err := service.DeleteObject(context.Background(), &bucket, &key)
		assert.Nil(t, err)
	})
}
