function keyStrings(str1, str2) {
    return str1 > str2 ? str1 + str2 : str2 + str1
}

class NetworkView extends React.Component {
    constructor(props) {
        super(props)
        this.refresh()
    }

    state = {
        graph: {}
    };

    refresh() {
        new NetworkService('.').getPeerGraph().then((graph) => {
            this.setState({graph: graph})
        })
    }

    getNodeLabel(id) {
        const parts = id.split("-")
        return parts[parts.length - 1]
    }

    componentDidUpdate() {
        const graph = this.state.graph

        const allEdges = graph.map((node) => node.peers.map(peer => ({from: node.id, to: peer}))).flat(1)
        // now allEdges contains logical duplicates because connections are bidirectional, which need to be filtered out:
        const uniqueEdges = Object.values(allEdges.reduce((map, obj) => {
            map[keyStrings(obj.from, obj.to)] = obj
            return map
        }, {}))

        let dot = "digraph {\n"
        dot += "  edge [\n" +
               "    arrowhead=\"none\"\n" +
               "  ];\n"
        const localLabel = this.getNodeLabel(graph.filter(node => node.self)[0].id);
        dot += "  \"" + localLabel + "\" [label=\"" + localLabel + "\\n(local)\"];\n"
        uniqueEdges.forEach(edge => { dot += "  \"" + this.getNodeLabel(edge.from) +  "\" -> \"" + this.getNodeLabel(edge.to) + "\";\n" })
        dot += "}"
        d3.select(this.el).graphviz().renderDot(dot)
    }

    render() {
        return <div ref={el => this.el = el}/>
    }
}