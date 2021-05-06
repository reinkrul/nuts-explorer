class DIDService {
    constructor(basePath) {
        this.basePath = basePath
    }

    list() {
        return fetch(this.basePath + "/api/vdr")
            .then((response) => response.json())
    }

    get(did) {
        return fetch(this.basePath + "/api/vdr/" + did)
            .then((response) => response.json())
    }
}