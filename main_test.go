package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ory/dockertest/v3"
)

const (
	localStackVersion = "2.0.2"
)

var dockerResource *dockertest.Resource

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("error creating dockertest pool: %s", err)
	}

	if err := pool.Client.Ping(); err != nil {
		log.Fatalf("error pinging pool: %s", err)
	}

	options := &dockertest.RunOptions{
		Repository:   "localstack/localstack",
		Tag:          localStackVersion,
		ExposedPorts: []string{"4566/tcp", "4572/tcp"},
	}

	resource, err := pool.RunWithOptions(options)
	if err != nil {
		log.Fatalf("error starting localstack: %s", err)
	}

	type hcResponse struct {
		Services struct {
			S3 string `json:"s3"`
		} `json:"services"`
	}

	if err := pool.Retry(func() error {
		healthCheckURL := fmt.Sprintf("http://%s/_localstack/health", net.JoinHostPort("localhost", resource.GetPort("4566/tcp")))

		resp, err := http.Get(healthCheckURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var response hcResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}

		if response.Services.S3 == "available" {
			return nil
		}

		return errors.New("S3 not available")
	}); err != nil {
		log.Fatalf("could not connect to localstack: %s", err)
	}

	dockerResource = resource

	exitCode := m.Run()

	// this can't be deferred, because of the os.Exit() call below.
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resources: %s", err)
	}

	os.Exit(exitCode)
}

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

// testS3Client creates an S3 client from the given AWS configuration, and configure it
// to use "Path Style" (i.e. http://s3.amazonaws.com/BUCKET-NAME/key); this is needed to use S3
// with localstack without having to create custom DNS entries for each bucket.
func testS3Client(cfg aws.Config) *s3.Client {
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

func TestBucketFileExists(t *testing.T) {
	hostPort := fmt.Sprintf("localhost:%s", dockerResource.GetPort("4566/tcp"))
	awsConfig := testConfig(hostPort)
	s3Client := testS3Client(awsConfig)

	prog := Program{S3Client: s3Client}
	ctx := context.Background()

	if err := prog.CreateBucket(ctx, bucket); err != nil {
		t.Fatal(err)
	}

	if err := prog.CreateBucketFile(ctx, bucket, filename, []byte("hello world")); err != nil {
		t.Fatal(err)
	}

	found, err := prog.BucketFileExists(ctx, bucket, filename)
	if err != nil {
		t.Fatal(err)
	}

	if !found {
		t.Fatal("file not found")
	}
}
