package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

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

func TestHeadObject(t *testing.T) {
	hostPort := fmt.Sprintf("localhost:%s", dockerResource.GetPort("4566/tcp"))
	awsConfig := testConfig(hostPort)

	if err := run(awsConfig); err != nil {
		t.Fatal(err)
	}
}
