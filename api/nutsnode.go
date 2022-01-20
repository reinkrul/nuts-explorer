/*
 * Copyright (C) 2021 Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package api

import (
	"encoding/json"
	"fmt"
	v1 "github.com/nuts-foundation/nuts-node/network/api/v1"
	"github.com/nuts-foundation/nuts-node/network/p2p"
	"io"
	"net/http"
	"regexp"
	"time"
)

type NutsNodeService struct {
	APIAddress    string
	StatusAddress string
}

func (g NutsNodeService) GetNetworkGraph(w http.ResponseWriter) error {
	graph, err := g.getNetworkGraph()
	if err != nil {
		return err
	}
	return respondOK(w, graph)
}

var ownPeerRegex = regexp.MustCompile("\\[P2P Network] Peer ID of local node: (.*)\n")

type node struct {
	ID    p2p.PeerID   `json:"id"`
	Self  bool         `json:"self"`
	Peers []p2p.PeerID `json:"peers"`
}

func (g NutsNodeService) getNetworkGraph() ([]node, error) {
	nodes := make(map[p2p.PeerID]node, 0)

	// Get local node
	ownPeerIDStr, err := g.getNodePeerID()
	if err != nil {
		return nil, err
	}
	ownPeerID := p2p.PeerID(ownPeerIDStr)

	// Register peer nodes
	peers, err := v1.HTTPClient{ServerAddress: g.APIAddress, Timeout: 5 * time.Second}.GetPeerDiagnostics()
	if err != nil {
		return nil, err
	}
	for peerID, diagnostics := range peers {
		nodes[peerID] = node{Peers: diagnostics.Peers, Self: peerID == ownPeerID}
	}

	// Register local node
	peerIDs := make([]p2p.PeerID, 0, len(peers))
	for p := range peers {
		peerIDs = append(peerIDs, p)
	}
	nodes[ownPeerID] = node{Self: true, Peers: peerIDs}

	// Convert to result slice
	var result []node
	for id, curr := range nodes {
		result = append(result, curr)
		i := len(result) - 1
		result[i].ID = id
		// nil slices break the frontend
		if len(result[i].Peers) == 0 {
			result[i].Peers = []p2p.PeerID{}
		}
	}
	return result, nil
}

func (g NutsNodeService) getNodePeerID() (string, error) {
	request, err := http.NewRequest("GET", g.StatusAddress+"/status/diagnostics", nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	diagnosticsBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{}
	err = json.Unmarshal(diagnosticsBytes, &data)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", (data["network"].(map[string]interface{}))["peer_id"]), nil
}

func respondOK(writer http.ResponseWriter, body interface{}) error {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	var data []byte
	if bodyAsBytes, ok := body.([]byte); ok {
		data = bodyAsBytes
	} else {
		if bodyAsBytes, err := json.Marshal(body); err != nil {
			return err
		} else {
			data = bodyAsBytes
		}
	}
	_, err := writer.Write(data)
	return err
}
