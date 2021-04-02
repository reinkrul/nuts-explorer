package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/nuts-foundation/go-did"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const jsonMimeType = "application/json"

type ServiceProxy struct {
	Address string
}

func (g ServiceProxy) ListDIDs(w http.ResponseWriter) error {
	transactions, err := g.getTransactions()
	if err != nil {
		return err
	}

	type entry struct {
		DID     string `json:"did"`
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
	w.Header().Add("Content-Type", jsonMimeType)
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(results)
	_, err = w.Write(data)
	return err
}

func (g ServiceProxy) SearchVCs(w http.ResponseWriter, concept string, query []byte) error {
	resp, err := http.Post(g.Address+"/internal/vcr/v1/"+url.PathEscape(concept), jsonMimeType, bytes.NewReader(query))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	results, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	w.Header().Add("Content-Type", jsonMimeType)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(results)
	return err
}

func (g ServiceProxy) getTransactions() ([]*jws.Message, error) {
	resp, err := http.Get(g.Address + "/api/transaction")
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

func (g ServiceProxy) ResolveDID(writer http.ResponseWriter, didToResolve string) error {
	if !strings.HasPrefix(didToResolve, "did:nuts:") {
		return fmt.Errorf("invalid DID to resolve: %s", didToResolve)
	}

	resp, err := http.Get(g.Address + "/internal/vdr/v1/did/" + didToResolve)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(data)
	return err
}
