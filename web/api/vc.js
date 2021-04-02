class VCService {
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

    getUntrustedVCs() {
        // TODO
        return [];
    }
}