package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	bucket   = "test"
	filename = "foo.bar"
)

func endpointResolver(service, region string, options ...any) (aws.Endpoint, error) {
	return aws.Endpoint{
		PartitionID:   "aws",
		URL:           "http://localhost:4566",
		SigningRegion: "us-east-1",
	}, nil
}

func run() error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(endpointResolver)),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("AKID", "SECRET_KEY", "TOKEN")),
	)
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// create the bucket
	if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}); err != nil {
		return err
	}

	// put a fake file
	if _, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   strings.NewReader("hello world"),
	}); err != nil {
		return err
	}

	// try to read the file
	info, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return err
	}

	log.Printf("info: %+v", info)

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %s", err)
	}
}
