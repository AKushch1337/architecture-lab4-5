package integration

import (
	"bytes"
	"fmt"
	. "gopkg.in/check.v1"
	"net/http"
	"os"
	"testing"
	"time"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

type IntegrationTestSuite struct{}

var _ = Suite(&IntegrationTestSuite{})

func TestBalancer(t *testing.T) {
	TestingT(t)
}

func (s *IntegrationTestSuite) TestGetRequest(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		resp, err = client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil {
			c.Error(err)
		}
		c.Check(resp.Header.Get("lb-from"), Equals, "server"+string(rune(i+1))+":8080")
	}
	resp, err = client.Post("http://server1:8080/corrupt-health", "", bytes.NewBuffer([]byte{}))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.Error(err)
	}
	time.Sleep(10 * time.Second)
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
	if err != nil {
		c.Error(err)
	}
	c.Check(resp.Header.Get("lb-from"), Equals, "server2:8080")
}

func (s *IntegrationTestSuite) BenchmarkBalancer(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}

	for i := 0; i < c.N; i++ {
		_, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil {
			c.Error(err)
		}
	}
}
