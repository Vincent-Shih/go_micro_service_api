// quite useless mock test, but it's a good example of how to mock AWS SDK for Go v2
package s3_test

import (
	"context"
	impl "go_micro_service_api/pkg/storage/s3"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}

func TestMockClient(t *testing.T) {
	cfg, err := impl.NewAWSConfig(context.Background())
	assert.Nil(t, err)

	client := impl.NewClient(cfg)
	assert.NotNil(t, client)
}

func TestMockService(t *testing.T) {
	client := new(MockS3Client)
	service := impl.NewService(client)
	assert.NotNil(t, service)

	bucket := "go_micro_service_api"
	prefix := "test/"
	key := "test/test.txt"
	body := "Hello, World!"
	t.Run("PutObject", func(t *testing.T) {
		versionId := "123"
		client.On("PutObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.PutObjectOutput{
			VersionId: &versionId,
		}, nil)
		result, err := service.PutObject(context.Background(), &bucket, &key, strings.NewReader(body))
		assert.Nil(t, err)

		assert.Equal(t, versionId, *result.VersionId)
	})

	t.Run("ListObjects", func(t *testing.T) {
		client.On("ListObjectsV2", mock.Anything, mock.Anything, mock.Anything).Return(&s3.ListObjectsV2Output{
			Contents: []types.Object{
				{
					Key: &key,
				},
			},
		}, nil)
		result, err := service.ListObjects(context.Background(), &bucket, &prefix, nil)
		assert.Nil(t, err)

		for _, obj := range result.Contents {
			assert.Equal(t, key, aws.ToString(obj.Key))
		}
	})

	t.Run("GetObject", func(t *testing.T) {
		client.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{
			Body: io.NopCloser(strings.NewReader(body)),
		}, nil)
		result, err := service.GetObject(context.Background(), &bucket, &key)
		assert.Nil(t, err)

		content, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, []byte(body), content)
	})

	t.Run("DeleteObject", func(t *testing.T) {
		client.On("DeleteObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.DeleteObjectOutput{}, nil)
		err := service.DeleteObject(context.Background(), &bucket, &key)
		assert.Nil(t, err)
	})
}
