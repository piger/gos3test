package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
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
		var responsError *awshttp.ResponseError
		if errors.As(err, &responsError) && responsError.HTTPStatusCode() == http.StatusNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
