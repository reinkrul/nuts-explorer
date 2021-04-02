class DIDService {
    list() {
        return fetch("/api/vdr")
            .then((response) => response.json())
    }

    get(did) {
        return fetch("/api/vdr/" + did)
            .then((response) => response.json())
    }
}