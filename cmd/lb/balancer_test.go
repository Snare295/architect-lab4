package main

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type BalancerRouterSuite struct{}

var _ = Suite(&BalancerRouterSuite{})

func (s *BalancerRouterSuite) TestCreateServersInsts(c *C) {
	b := &BalancerRouter{}
	serversPool := []string{"server1:8080", "server2:8080", "server3:8080"}
	b.createServersInsts(serversPool)
	c.Assert(len(b.servers), Equals, len(serversPool))
}

func (s *BalancerRouterSuite) TestFindServerByUrl(c *C) {
	b := &BalancerRouter{}
	b.servers = []ServerLoad{
		{serverName: "server1"},
		{serverName: "server2"},
		{serverName: "server3"},
	}
	b.addNewLiveServer("server1", true)
	b.addNewLiveServer("server2", true)
	b.addNewLiveServer("server3", true)

	serverIndex := b.findServerByUrl("guy")
	c.Assert(serverIndex, Equals, 2)

	serverIndex = b.findServerByUrl("others")
	c.Assert(serverIndex, Equals, 0)

	serverIndex = b.findServerByUrl("aurelious1")
	c.Assert(serverIndex, Equals, 1)

	serverIndex = b.findServerByUrl("hello")
	c.Assert(serverIndex, Equals, 1)
}

// func (s *BalancerRouterSuite) TestPutClientToServ(c *C) {
// 	b := &BalancerRouter{}
// 	b.servers = []ServerLoad{
// 		{serverName: "server1"},
// 		{serverName: "server2"},
// 		{serverName: "server3"},
// 	}
// 	b.liveServers = []*ServerLoad{
// 		&b.servers[0], &b.servers[1], &b.servers[2],
// 	}

// 	err, serverIndex := b.putClientToServ("/newpath")
// 	c.Assert(err, IsNil)
// 	c.Assert(b.findServerByUrl("/newpath"), Equals, serverIndex)

// 	err, serverIndex = b.putClientToServ("/newpath2")
// 	c.Assert(err, IsNil)
// 	c.Assert(b.findServerByUrl("/newpath2"), Equals, serverIndex)

// 	err, serverIndex = b.putClientToServ("/newpath3")
// 	c.Assert(err, IsNil)
// 	c.Assert(b.findServerByUrl("/newpath3"), Equals, serverIndex)
// }

// func (s *BalancerRouterSuite) TestRemoveClientsFromDead(c *C) {
// 	b := &BalancerRouter{}
// 	b.servers = []ServerLoad{
// 		{serverName: "server1", conHash: []uint16{100, 200}},
// 		{serverName: "server2", conHash: []uint16{300, 400}},
// 		{serverName: "server3", conHash: []uint16{500, 600}},
// 	}

// 	b.removeClientsFromDead("server2")
// 	c.Assert(len(b.servers[1].conHash), Equals, 0)

// 	b.removeClientsFromDead("server3")
// 	c.Assert(len(b.servers[2].conHash), Equals, 0)

// 	b.removeClientsFromDead("server1")
// 	c.Assert(len(b.servers[0].conHash), Equals, 0)
// }

func (s *BalancerRouterSuite) TestAddNewLiveServer(c *C) {
	b := &BalancerRouter{}
	b.servers = []ServerLoad{
		{serverName: "server1"},
		{serverName: "server2"},
		{serverName: "server3"},
	}

	b.addNewLiveServer("server2", true)
	c.Assert(len(b.liveServers), Equals, 1)
	c.Assert(b.liveServers[0].serverName, Equals, "server2")

	b.addNewLiveServer("server1", true)
	c.Assert(len(b.liveServers), Equals, 2)
	c.Assert(b.liveServers[1].serverName, Equals, "server1")

	b.addNewLiveServer("server1", false)
	c.Assert(len(b.liveServers), Equals, 1)
	c.Assert(b.liveServers[0].serverName, Equals, "server2")
}

func (s *BalancerRouterSuite) TestScheme(c *C) {
	*https = true
	c.Assert(scheme(), Equals, "https")

	*https = false
	c.Assert(scheme(), Equals, "http")
}
