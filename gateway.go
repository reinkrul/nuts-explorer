package main

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
)

const jsonMimeType = "application/json"

type serviceProxy struct {
	address string
}

func (g serviceProxy) ListDIDs(w http.ResponseWriter) error {
	transactions, err := g.getTransactions()
	if err != nil {
		return err
	}

	resultsAsMap := make(map[string]bool, 0)
	var results []string
	for _, tx := range transactions {
		hdrs := tx.Signatures()[0].ProtectedHeaders()
		if hdrs.ContentType() == "application/did+json" {
			// This is a DID. Derive DID from JWK key ID.
			var keyID *did.DID
			if hdrs.JWK() != nil {
				if keyID, err = did.ParseDID(hdrs.JWK().KeyID()); err != nil {
					return err
				}
			} else {
				if keyID, err = did.ParseDID(hdrs.KeyID()); err != nil {
					return err
				}
			}
			curr := *keyID
			curr.Fragment = ""
			// Updates end up as duplicate entries
			if !resultsAsMap[curr.String()] {
				results = append(results, curr.String())
			}
			resultsAsMap[curr.String()] = true
		}
	}
	w.Header().Add("Content-Type", jsonMimeType)
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(results)
	_, err = w.Write(data)
	return err
}


func (g serviceProxy) SearchVCs(w http.ResponseWriter, concept string, query []byte) error {
	resp, err := http.Post(g.address + "/internal/vcr/v1/" + url.PathEscape(concept), jsonMimeType, bytes.NewReader(query))
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

func (g serviceProxy) getTransactions() ([]*jws.Message, error) {
	resp, err := http.Get(g.address + "/api/transaction")
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

func (g serviceProxy) ResolveDID(writer http.ResponseWriter, didToResolve string) error {
	if !strings.HasPrefix(didToResolve, "did:nuts:") {
		return fmt.Errorf("invalid DID to resolve: %s", didToResolve)
	}

	resp, err := http.Get(g.address + "/internal/vdr/v1/did/" + didToResolve)
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
