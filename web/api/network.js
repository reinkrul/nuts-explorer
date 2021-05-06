class NetworkService {
    constructor(basePath) {
        this.basePath = basePath
    }

    getPeerGraph() {
        return fetch(this.basePath + "/api/network/peergraph").then(response => response.json())
    }

    getDAG() {
        return fetch(this.basePath + "/api/network/dag").then(response => response.text())
    }
}