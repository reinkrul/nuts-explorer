package main

import (
	"embed"
	"fmt"
	"github.com/gorilla/mux"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

const EnvAddr = "NUTS_NODE_ADDRESS"

//go:embed web/*
var app embed.FS

func main() {
	var serverAddr = ":8080"
	log.Println("Starting server on", serverAddr)
	nutsNodeAddress := os.Getenv(EnvAddr)
	if nutsNodeAddress == "" {
		panic(EnvAddr + " not set")
	}
	nutsNodeAddress = strings.TrimSuffix(nutsNodeAddress, "/")
	log.Println("Proxying calls to Nuts Node on", nutsNodeAddress)

	router := mux.NewRouter()
	registerWebApp(router)
	registerAPI(router, serviceProxy{address: nutsNodeAddress})
	http.Handle("/", router)
	_ = http.ListenAndServe(serverAddr, nil)
}

func registerAPI(router *mux.Router, proxy serviceProxy) {
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
}

func registerWebApp(router *mux.Router) {
	var sysFS fs.FS
	if os.Getenv("LIVE_MODE") == "true" {
		sysFS = os.DirFS("web")
	} else {
		sysFS, _ = fs.Sub(app, "web")
	}
	webapp := http.FileServer(http.FS(sysFS))
	router.Handle("/", webapp)
}

func sendError(writer http.ResponseWriter, request *http.Request, err error) {
	log.Println("ERROR:", err)
	writer.Header().Add("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusInternalServerError)
	_, _ = writer.Write([]byte(fmt.Sprintf("unable to handle request to %s: %v", request.RequestURI, err.Error())))
}
