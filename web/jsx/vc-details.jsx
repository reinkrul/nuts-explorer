class VCDetails extends React.Component {
    state = {
        resolutionResult: {
            document: {}
        }
    }

    componentDidMount() {
        new VCService().get(this.props.id).then(doc => {
            this.setState({resolutionResult: doc})
        })
    }

    render() {
        return <div>
            <h2>VC: <code>{this.props.id}</code></h2>
            <pre><code>{JSON.stringify(this.state.resolutionResult.document, null, 2)}</code></pre>
        </div>
    }
}