class DIDDetails extends React.Component {
    state = {
        resolutionResult: {
            document: {}
        }
    }

    componentDidMount() {
        new DIDService().get(this.props.did).then(doc => {
            this.setState({resolutionResult: doc})
        })
    }

    render() {
        return <div>
            <h2>DID: <code>{this.props.did}</code></h2>
            <pre><code>{JSON.stringify(this.state.resolutionResult.document, null, 2)}</code></pre>
        </div>
    }
}