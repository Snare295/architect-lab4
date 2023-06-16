package integration
import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 2 * time.Second,
}

func (s *MySuite) TestBalancer(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Test of integration are not available")
	}

	// Testing servers
	server1, err := client.Get(fmt.Sprintf("%s/check", baseAddress))
	c.Assert(err, IsNil, Commentf("Request to serv1 don't sends : %v", err))
	server2, err := client.Get(fmt.Sprintf("%s/check4", baseAddress))
	c.Assert(err, IsNil, Commentf("Request to serv2 don't sends: %v", err))
	server3, err := client.Get(fmt.Sprintf("%s/check2", baseAddress))
	c.Assert(err, IsNil, Commentf("Request to serv3 don't sends: %v", err))
	server1Repeat, err := client.Get(fmt.Sprintf("%s/check", baseAddress))
	c.Assert(err, IsNil, Commentf("Sending repeated request to server1: %v", err))

	// Testing headers of servers
	server1Header := server1.Header.Get("lb-from")
	c.Assert(server1Header, Equals, "server1:8080", Commentf("Wrong server for server1 - %s", server1Header))
	server2Header := server2.Header.Get("lb-from")
	c.Assert(server2Header, Equals, "server1:8080", Commentf("Wrong server for server2 - %s", server2Header))
	server3Header := server3.Header.Get("lb-from")
	c.Assert(server3Header, Equals, "server1:8080", Commentf("Wrong server for server3 - %s", server3Header))
	server1RepeatHeader := server1Repeat.Header.Get("lb-from")
	c.Assert(server1RepeatHeader, Equals, server1Header, Commentf("Headers are not equal. origin - %s, repeat - %s", server1Header, server1RepeatHeader))
}

func (s *MySuite) BenchmarkBalancer(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Test of integration are not available")
	}

	for i := 0; i < c.N; i++ {
		_, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil {	c.Error(err) }
	}
}