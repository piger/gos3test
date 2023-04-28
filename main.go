package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	bucket   = "test"
	filename = "foo.bar"
)

func run() error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(cfg)

	prog := Program{S3Client: client}

	if err := prog.CreateBucket(ctx, bucket); err != nil {
		return err
	}

	if err := prog.CreateBucketFile(ctx, bucket, filename, []byte("hello world")); err != nil {
		return err
	}

	found, err := prog.BucketFileExists(ctx, bucket, filename)
	if err != nil {
		return err
	}

	fmt.Printf("found file? %v", found)

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %s", err)
	}
}
