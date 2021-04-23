package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/nuts-foundation/go-did"
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

type ServiceProxy struct {
	APIAddress    string
	StatusAddress string
}

func (g ServiceProxy) GetNetworkGraph(w http.ResponseWriter) error {
	graph, err := g.getNetworkGraph()
	if err != nil {
		return err
	}
	return respondOK(w, graph)
}

func (g ServiceProxy) ListDIDs(w http.ResponseWriter) error {
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
	var results []*entry
	for _, tx := range transactions {
		hdrs := tx.Signatures()[0].ProtectedHeaders()
		if hdrs.ContentType() == "application/did+json" {
			// This is a DID. Derive DID from JWK key ID.
			var keyID *did.DID
			if hdrs.JWK() != nil {
				// New DID
				if keyID, err = did.ParseDID(hdrs.JWK().KeyID()); err != nil {
					return err
				}
			} else {
				// Update
				if keyID, err = did.ParseDID(hdrs.KeyID()); err != nil {
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

func (g ServiceProxy) SearchVCs(w http.ResponseWriter, concept string, query []byte) error {
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

func (g ServiceProxy) ResolveDID(writer http.ResponseWriter, didToResolve string) error {
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

func (g ServiceProxy) GetVC(writer http.ResponseWriter, id string) error {
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

func (g ServiceProxy) ListTrustedVCIssuers(writer http.ResponseWriter) error {
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

func (g ServiceProxy) ListUntrustedVCIssuers(writer http.ResponseWriter) error {
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

func (g ServiceProxy) GetDAG(writer http.ResponseWriter) error {
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
var peersRegex = regexp.MustCompile("\\[P2P Network] Connected peers: (.*)\n")

type node struct {
	ID    string   `json:"id"`
	Self  bool     `json:"self"`
	Peers []string `json:"peers"`
}

func (g ServiceProxy) getNetworkGraph() ([]node, error) {
	resp, err := http.Get(g.StatusAddress + "/status/diagnostics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	diagnosticsBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	diagnostics := string(diagnosticsBytes)

	// Own peer ID
	ownPeerIDStr := ownPeerRegex.FindStringSubmatch(diagnostics)
	if len(ownPeerIDStr) != 2 {
		return nil, errors.New("unable to find own peer ID in diagnostics")
	}
	ownPeerID := ownPeerIDStr[1]

	nodes := make(map[string]node, 0)
	nodes[ownPeerID] = node{
		ID:    ownPeerID,
		Self:  true,
		Peers: []string{},
	}
	// Peers
	peersStr := peersRegex.FindStringSubmatch(diagnostics)
	if len(peersStr) == 2 {
		peers := strings.Split(peersStr[1], " ")
		for _, peer := range peers {
			peerParts := strings.Split(peer, "@")
			if len(peerParts) == 2 {
				self := nodes[ownPeerID]
				peerID := peerParts[0]
				self.Peers = append(self.Peers, peerID)
				nodes[self.ID] = self
				nodes[peerID] = node{ID: peerID, Peers: []string{}}
			}
		}
	}

	result := []node{}
	for _, curr := range nodes {
		result = append(result, curr)
	}
	return result, nil
}

func (g ServiceProxy) getTransactions() ([]*jws.Message, error) {
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
