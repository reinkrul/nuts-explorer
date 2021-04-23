class NetworkService {
    getPeerGraph() {
        return fetch("/api/network/peergraph").then(response => response.json())
    }

    getDAG() {
        return fetch("/api/network/dag").then(response => response.text())
    }
}