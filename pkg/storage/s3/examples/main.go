package main

import (
	"context"
	"go_micro_service_api/pkg/storage/s3"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func main() {
	ctx := context.Background()
	cfg, err := s3.NewAWSConfig(ctx)
	if err != nil {
		panic(err)
	}
	client := s3.NewClient(cfg)
	service := s3.NewService(client)

	bucket := "go_micro_service_api-test"
	prefix := "test/"
	key := "test/test.txt"

	// PutObject
	putResult, err := service.PutObject(ctx, &bucket, &key, strings.NewReader("Hello, World!"))
	if err != nil {
		panic(err)
	}
	println(putResult.VersionId)

	// ListObjects
	listResult, err := service.ListObjects(ctx, &bucket, &prefix, nil)
	if err != nil {
		panic(err)
	}
	for _, obj := range listResult.Contents {
		println(aws.ToString(obj.Key), obj.Size)
	}

	// GetObject
	getResult, err := service.GetObject(ctx, &bucket, &key)
	if err != nil {
		panic(err)
	}
	defer getResult.Body.Close()
	println(io.ReadAll(getResult.Body))

	// DeleteObject
	err = service.DeleteObject(ctx, &bucket, &key)
	if err != nil {
		panic(err)
	}
}
