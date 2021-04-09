class NetworkService {
    getGraph() {
        return fetch("/api/network/graph").then(response => response.json())
    }
}