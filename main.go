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

package main

import (
	"embed"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nuts-foundation/nuts-did-explorer/api"
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
	registerAPI(router, api.NutsNodeService{APIAddress: nutsNodeAddress, StatusAddress: nutsNodeStatusAddress})
	registerWebApp(router)
	http.Handle("/", router)
	err := http.ListenAndServe(serverAddr, nil)
	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Unable to start listener: %v", err))
	}
}

func registerAPI(router *mux.Router, proxy api.NutsNodeService) {
	router.HandleFunc("/api/network/peergraph", func(writer http.ResponseWriter, request *http.Request) {
		if err := proxy.GetNetworkGraph(writer); err != nil {
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
