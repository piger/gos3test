package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	localStackVersion = "2.0.2"
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("error creating dockertest pool: %s", err)
	}

	if err := pool.Client.Ping(); err != nil {
		log.Fatalf("error pinging pool: %s", err)
	}

	options := &dockertest.RunOptions{
		Repository: "localstack/localstack",
		Tag:        localStackVersion,
		PortBindings: map[docker.Port][]docker.PortBinding{
			"4566/tcp": {
				{
					HostPort: "4566",
				},
			},
			"4572/tcp": {
				{
					HostPort: "4572",
				},
			},
		},
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
		resp, err := http.Get("http://127.0.0.1:4566/_localstack/health")
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

	exitCode := m.Run()

	// this can't be deferred, because of the os.Exit() call below.
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resources: %s", err)
	}

	os.Exit(exitCode)
}

func TestHeadObject(t *testing.T) {
	if err := run(); err != nil {
		t.Fatal(err)
	}
}
