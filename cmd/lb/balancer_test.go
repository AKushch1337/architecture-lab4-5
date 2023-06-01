package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jarcoal/httpmock"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) SetUpSuite(c *C) {
	httpmock.Activate()
}

func (s *MySuite) TearDownSuite(c *C) {
	httpmock.DeactivateAndReset()
}

func (s *MySuite) TestFindMinServer(c *C) {
	serversPool = nil

	serversPool = []*Server{
		{URL: "Server1", ConnCnt: 10, Healthy: false},
		{URL: "Server2", ConnCnt: 20, Healthy: false},
		{URL: "Server3", ConnCnt: 30, Healthy: false},
	}
	c.Assert(minServerIndex(), Equals, -1)

	serversPool = []*Server{
		{URL: "Server1", ConnCnt: 10, Healthy: true},
		{URL: "Server2", ConnCnt: 20, Healthy: true},
		{URL: "Server3", ConnCnt: 30, Healthy: true},
	}
	c.Assert(minServerIndex(), Equals, 0)

	serversPool = []*Server{
		{URL: "Server1", ConnCnt: 10, Healthy: false},
		{URL: "Server2", ConnCnt: 20, Healthy: true},
		{URL: "Server3", ConnCnt: 30, Healthy: true},
	}
	c.Assert(minServerIndex(), Equals, 1)

	serversPool = []*Server{
		{URL: "Server1", ConnCnt: 10, Healthy: true},
		{URL: "Server2", ConnCnt: 5, Healthy: true},
		{URL: "Server3", ConnCnt: 30, Healthy: true},
	}
	c.Assert(minServerIndex(), Equals, 1)
}

func (s *MySuite) TestHealth(c *C) {
	mockURL := "http://example.com/health"
	httpmock.RegisterResponder(http.MethodGet, mockURL, httpmock.NewStringResponder(http.StatusOK, ""))

	server := &Server{
		URL: "example.com",
	}

	result := health(server)

	c.Assert(result, Equals, true)
	c.Assert(server.Healthy, Equals, true)

	server.Healthy = false

	httpmock.RegisterResponder(http.MethodGet, mockURL, httpmock.NewStringResponder(http.StatusInternalServerError, ""))
	result2 := health(server)

	c.Assert(result2, Equals, false)
	c.Assert(server.Healthy, Equals, false)
}

func (s *MySuite) TestForward(c *C) {
	httpmock.RegisterResponder("GET", "http://server1:8080/",
		httpmock.NewStringResponder(200, "OK"))

	serversPool = []*Server{
		{URL: "server1:8080", Healthy: true},
	}

	req, err := http.NewRequest("GET", "/", nil)
	c.Assert(err, IsNil)
	rr := httptest.NewRecorder()
	err = forward(rr, req)
	c.Assert(err, IsNil)
}

func (s *MySuite) TestForwardWithUnhealthyServer(c *C) {
	httpmock.RegisterResponder("GET", "http://server1:8080/",
		httpmock.NewStringResponder(500, "Error"))

	serversPool = []*Server{
		{URL: "server1:8080", Healthy: false},
	}

	req, err := http.NewRequest("GET", "/", nil)
	c.Assert(err, IsNil)
	rr := httptest.NewRecorder()
	err = forward(rr, req)
	c.Assert(err, NotNil)
}
