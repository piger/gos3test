package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	bucket   = "test"
	filename = "foo.bar2"
)

type Configurer func() aws.Config

// testConfig creates an AWS configuration using test credentials and an endpoint resolver
// that points to localstack on localhost.
func testConfig(hostPort string) aws.Config {
	endpointResolver := func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           fmt.Sprintf("http://%s", hostPort),
			SigningRegion: "us-east-1",
		}, nil
	}

	return aws.Config{
		Region:                      "us-east-1",
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(endpointResolver),
		Credentials:                 credentials.NewStaticCredentialsProvider("AKID", "SECRET_KEY", "TOKEN"),
	}
}

func normalConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	return cfg
}

// testS3Client creates an S3 client from the given AWS configuration, and configure it
// to use "Path Style" (i.e. http://s3.amazonaws.com/BUCKET-NAME/key); this is needed to use S3
// with localstack without having to create custom DNS entries for each bucket.
func testS3Client(cfg aws.Config) *s3.Client {
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

func run(cfg aws.Config) error {
	ctx := context.Background()

	client := testS3Client(cfg)

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
	_, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(normalConfig()); err != nil {
		log.Fatalf("error: %s", err)
	}
}
