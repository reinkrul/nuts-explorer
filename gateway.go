package main

import (
	"encoding/json"
	"fmt"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/nuts-foundation/go-did"
	"io"
	"net/http"
	"strings"
)

type serviceProxy struct {
	address string
}

func (g serviceProxy) ListDIDs(w http.ResponseWriter) error {
	resp, err := http.Get(g.address + "/api/transaction")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	transactionsAsBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var transactions []string
	if err := json.Unmarshal(transactionsAsBytes, &transactions); err != nil {
		return err
	}
	resultsAsMap := make(map[string]bool, 0)
	var results []string
	for _, str := range transactions {
		if tx, err := jws.ParseString(str); err != nil {
			return err
		} else {
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
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(results)
	_, err = w.Write(data)
	return err
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
