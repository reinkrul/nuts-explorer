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

    getNodeLabel(node) {
        const parts = node.id.split("-")
        return parts[parts.length - 1]
    }

    componentDidUpdate() {
        const graph = this.state.graph
        const edges = graph.map((node) => node.peers.map(peer => ({from: node.id, to: peer}))).flat(1)
        const nodes = graph.map(node => ({
            id: node.id,
            label: this.getNodeLabel(node),
            color: node.self ? 'lightblue' : 'lightgray',
        }))
        new vis.Network(this.el, {
            nodes: new vis.DataSet(nodes),
            edges: new vis.DataSet(edges)
        }, {
            height: '600',
            width: '700'
        })
    }

    render() {
        return <div ref={el => this.el = el}/>
    }
}