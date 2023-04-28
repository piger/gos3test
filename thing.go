package main

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Program struct {
	S3Client *s3.Client
}

func (p *Program) CreateBucket(ctx context.Context, name string) error {
	if _, err := p.S3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}); err != nil {
		return err
	}

	return nil
}

func (p *Program) CreateBucketFile(ctx context.Context, bucket, filename string, contents []byte) error {
	if _, err := p.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(contents),
	}); err != nil {
		return err
	}
	return nil
}

func (p *Program) BucketFileExists(ctx context.Context, bucket, filename string) (bool, error) {
	if _, err := p.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	}); err != nil {
		return false, err
	}

	// assume no error from HeadOjbect means that the object exists; ideally the error
	// checking above would differentiate between 403 and 404.
	return true, nil
}
