package api

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestNutsNodeService_GetNetworkGraph(t *testing.T) {
	t.Skip("uses specific external node address")
	service := NutsNodeService{
		APIAddress:    "http://rd9djud2-mgmt.nnaas.reinkrul.nl",
		StatusAddress: "http://8vfgx1g1.nnaas.reinkrul.nl",
	}
	graph, err := service.getNetworkGraph()
	if !assert.NoError(t, err) {
		return
	}
	for _, node := range graph {
		log.Printf("%v\n", node)
	}
}

func TestNutsNodeService_getNodePeerID(t *testing.T) {
	t.Skip("uses specific external node address")
	service := NutsNodeService{
		APIAddress:    "http://rd9djud2-mgmt.nnaas.reinkrul.nl",
		StatusAddress: "http://8vfgx1g1.nnaas.reinkrul.nl",
	}
	peerID, err := service.getNodePeerID()
	if !assert.NoError(t, err) {
		return
	}
	assert.NotEmpty(t, peerID)
}
