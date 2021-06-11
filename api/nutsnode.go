package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/nuts-foundation/go-did/did"
	v1 "github.com/nuts-foundation/nuts-node/network/api/v1"
	"github.com/nuts-foundation/nuts-node/network/p2p"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const jsonMimeType = "application/json"

var vcTypes = []string{
	"NutsOrganizationCredential",
}

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

func (g NutsNodeService) ListDIDs(w http.ResponseWriter) error {
	transactions, err := g.getTransactions()
	if err != nil {
		return err
	}

	type entry struct {
		DID     string    `json:"did"`
		Created time.Time `json:"created"`
		Updated time.Time `json:"updated"`
	}
	resultsAsMap := make(map[string]*entry, 0)
	results := make([]*entry, len(resultsAsMap))
	for _, tx := range transactions {
		hdrs := tx.Signatures()[0].ProtectedHeaders()
		if hdrs.ContentType() == "application/did+json" {
			// This is a DID. Derive DID from JWK key ID.
			var keyID *did.DID
			if hdrs.JWK() != nil {
				// New DID
				if keyID, err = did.ParseDIDURL(hdrs.JWK().KeyID()); err != nil {
					return err
				}
			} else {
				// Update
				if keyID, err = did.ParseDIDURL(hdrs.KeyID()); err != nil {
					return err
				}
			}
			currDID := *keyID
			currDID.Fragment = ""
			currDIDStr := currDID.String()

			_, ok := resultsAsMap[currDIDStr]
			sigtInterf, _ := hdrs.Get("sigt")
			sigt := time.Unix(int64(sigtInterf.(float64)), 0)
			if ok {
				// Update
				if sigt.After(resultsAsMap[currDID.String()].Updated) {
					resultsAsMap[currDID.String()].Updated = sigt
				}
				if sigt.Before(resultsAsMap[currDID.String()].Created) {
					resultsAsMap[currDID.String()].Created = sigt
				}
			} else {
				// New DID
				resultsAsMap[currDIDStr] = &entry{
					DID:     currDIDStr,
					Created: sigt,
					Updated: sigt,
				}
				results = append(results, resultsAsMap[currDIDStr])
			}
		}
	}
	return respondOK(w, results)
}

func (g NutsNodeService) SearchVCs(w http.ResponseWriter, concept string, query []byte) error {
	resp, err := http.Post(g.APIAddress+"/internal/vcr/v1/"+url.PathEscape(concept), jsonMimeType, bytes.NewReader(query))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	results, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return respondOK(w, results)
}

func (g NutsNodeService) ResolveDID(writer http.ResponseWriter, didToResolve string) error {
	if !strings.HasPrefix(didToResolve, "did:nuts:") {
		return fmt.Errorf("invalid DID to resolve: %s", didToResolve)
	}

	resp, err := http.Get(g.APIAddress + "/internal/vdr/v1/did/" + didToResolve)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return respondOK(writer, data)
}

func (g NutsNodeService) GetVC(writer http.ResponseWriter, id string) error {
	if strings.Contains(id, "#") {
		id = url.PathEscape(id)
	}
	resp, err := http.Get(g.APIAddress + "/internal/vcr/v1/vc/" + id)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return respondOK(writer, data)
}

func (g NutsNodeService) ListTrustedVCIssuers(writer http.ResponseWriter) error {
	result := make(map[string][]string, 0)
	for _, vcType := range vcTypes {
		resp, err := http.Get(g.APIAddress + "/internal/vcr/v1/" + vcType + "/trusted")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var items []string
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		result[vcType] = items
	}
	return respondOK(writer, result)
}

func (g NutsNodeService) ListUntrustedVCIssuers(writer http.ResponseWriter) error {
	result := make(map[string][]string, 0)
	for _, vcType := range vcTypes {
		resp, err := http.Get(g.APIAddress + "/internal/vcr/v1/" + vcType + "/untrusted")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var items []string
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		result[vcType] = items
	}
	return respondOK(writer, result)
}

func (g NutsNodeService) GetDAG(writer http.ResponseWriter) error {
	resp, err := http.Get(g.APIAddress + "/internal/network/v1/diagnostics/graph")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	writer.Header().Add("Content-Type", "text/vnd.graphviz")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(data)
	return err
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
	resp, err := http.Get(g.StatusAddress + "/status/diagnostics")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	diagnosticsBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	diagnostics := string(diagnosticsBytes)

	// Own peer ID
	ownPeerIDStr := ownPeerRegex.FindStringSubmatch(diagnostics)
	if len(ownPeerIDStr) != 2 {
		return "", fmt.Errorf("unable to find own peer ID in diagnostics")
	}
	return ownPeerIDStr[1], nil
}

func (g NutsNodeService) getTransactions() ([]*jws.Message, error) {
	resp, err := http.Get(g.APIAddress + "/internal/network/v1/transaction")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	transactionsAsBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var transactionsAsStrings []string
	if err := json.Unmarshal(transactionsAsBytes, &transactionsAsStrings); err != nil {
		return nil, err
	}
	var transactions []*jws.Message
	for _, str := range transactionsAsStrings {
		if tx, err := jws.ParseString(str); err != nil {
			return nil, err
		} else {
			transactions = append(transactions, tx)
		}
	}
	return transactions, nil
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
