class DAGView extends React.Component {
    constructor(props) {
        super(props)
        this.refresh()
    }

    state = {
        graph: {}
    };

    refresh() {
        new NetworkService().getDAG().then((graph) => {
            this.setState({graph: graph})
        })
    }

    componentDidUpdate() {
        d3.select(this.el).graphviz().renderDot(this.state.graph)
    }

    render() {
        return <div ref={el => this.el = el}/>
    }
}