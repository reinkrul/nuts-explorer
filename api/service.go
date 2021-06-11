package api

import "net/http"

type Service interface {
	GetNetworkGraph(w http.ResponseWriter) error
	ListDIDs(w http.ResponseWriter) error
	SearchVCs(w http.ResponseWriter, concept string, query []byte) error
	ResolveDID(writer http.ResponseWriter, didToResolve string) error
	GetVC(writer http.ResponseWriter, id string) error
	ListTrustedVCIssuers(writer http.ResponseWriter) error
	ListUntrustedVCIssuers(writer http.ResponseWriter) error
	GetDAG(writer http.ResponseWriter) error
}
