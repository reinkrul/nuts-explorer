const VCTypes = [
    "organization"
]

class VCService {
    constructor(basePath) {
        this.basePath = basePath
    }

    search(concept, query) {
        const params = Object.entries(query).map((kv) => ({
            key: [concept, kv[0]].join('.'),
            value: kv[1],
        }))
        return fetch("/api/vcr/search/" + concept, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({params: params})
        }).then(response => response.json())
    }

    get(id) {
        return fetch(this.basePath + "/api/vcr/" + encodeURIComponent(id)).then((response) => response.json())
    }

    getUntrustedIssuers() {
        return fetch(this.basePath + "/api/vcr/untrusted").then((response) => response.json())
    }

    getTrustedIssuers() {
        return fetch(this.basePath + "/api/vcr/trusted").then((response) => response.json())
    }
}