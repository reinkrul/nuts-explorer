package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nuts-foundation/nuts-did-explorer/api"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

const EnvListenAddr = "NUTS_EXPLORER_ADDRESS"
const EnvAddr = "NUTS_NODE_ADDRESS"
const EnvStatusAddr = "NUTS_NODE_STATUS_ADDRESS"

//go:embed web
var app embed.FS

func main() {
	var serverAddr string
	if serverAddr = os.Getenv(EnvListenAddr); serverAddr == "" {
		serverAddr = ":8080"
	}
	log.Println("Starting server on", serverAddr)
	nutsNodeAddress := os.Getenv(EnvAddr)
	if nutsNodeAddress == "" {
		panic(EnvAddr + " not set")
	}
	nutsNodeStatusAddress := os.Getenv(EnvStatusAddr)
	if nutsNodeStatusAddress == "" {
		nutsNodeStatusAddress = nutsNodeAddress
	}
	nutsNodeAddress = strings.TrimSuffix(nutsNodeAddress, "/")
	log.Println("Proxying calls to Nuts Node on", nutsNodeAddress)

	router := mux.NewRouter()
	registerAPI(router, api.ServiceProxy{APIAddress: nutsNodeAddress, StatusAddress: nutsNodeStatusAddress})
	registerWebApp(router)
	http.Handle("/", router)
	err := http.ListenAndServe(serverAddr, nil)
	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Unable to start listener: %v", err))
	}
}

func registerAPI(router *mux.Router, proxy api.ServiceProxy) {
	router.HandleFunc("/api/vdr", func(writer http.ResponseWriter, request *http.Request) {
		if err := proxy.ListDIDs(writer); err != nil {
			sendError(writer, request, err)
		}
	})
	router.HandleFunc("/api/vdr/{did}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		if err := proxy.ResolveDID(writer, vars["did"]); err != nil {
			sendError(writer, request, err)
		}
	})
	router.HandleFunc("/api/vcr/search/{concept}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		query, err := io.ReadAll(request.Body)
		if err != nil {
			sendError(writer, request, err)
			return
		}
		// Make sure the query is valid JSON
		var queryAsMap map[string]interface{}
		if json.Unmarshal(query, &queryAsMap) != nil {
			sendError(writer, request, errors.New("VC search query isn't valid JSON"))
			return
		}

		if err := proxy.SearchVCs(writer, vars["concept"], query); err != nil {
			sendError(writer, request, err)
		}
	})
	router.HandleFunc("/api/vcr/untrusted", func(writer http.ResponseWriter, request *http.Request) {
		if err := proxy.ListUntrustedVCIssuers(writer); err != nil {
			sendError(writer, request, err)
		}
	})
	router.HandleFunc("/api/vcr/trusted", func(writer http.ResponseWriter, request *http.Request) {
		if err := proxy.ListTrustedVCIssuers(writer); err != nil {
			sendError(writer, request, err)
		}
	})
	router.HandleFunc("/api/vcr/{id}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		if err := proxy.GetVC(writer, vars["id"]); err != nil {
			sendError(writer, request, err)
		}
	})
	router.HandleFunc("/api/network/peergraph", func(writer http.ResponseWriter, request *http.Request) {
		if err := proxy.GetNetworkGraph(writer); err != nil {
			sendError(writer, request, err)
		}
	})
	router.HandleFunc("/api/network/dag", func(writer http.ResponseWriter, request *http.Request) {
		if err := proxy.GetDAG(writer); err != nil {
			sendError(writer, request, err)
		}
	})
}

func registerWebApp(router *mux.Router) {
	var sysFS fs.FS
	if os.Getenv("LIVE_MODE") == "true" {
		sysFS = os.DirFS("web")
	} else {
		sysFS, _ = fs.Sub(app, "web")
	}
	webapp := http.FileServer(http.FS(sysFS))
	router.PathPrefix("/").Handler(webapp)
}

func sendError(writer http.ResponseWriter, request *http.Request, err error) {
	log.Println("ERROR:", err)
	writer.Header().Add("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusInternalServerError)
	_, _ = writer.Write([]byte(fmt.Sprintf("unable to handle request to %s: %v", request.RequestURI, err.Error())))
}
