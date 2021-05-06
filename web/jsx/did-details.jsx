class DIDDetails extends React.Component {
    constructor(props) {
        super(props)
        this.refresh();
    }

    state = {
        resolutionResult: {
            document: {}
        }
    }

    refresh() {
        new DIDService('.').get(this.props.did).then(doc => {
            this.setState({resolutionResult: doc})
        })
    }

    componentDidUpdate(prevProps) {
        if (prevProps.did !== this.props.did) {
            this.refresh();
        }
    }

    render() {
        return <div>
            <h2>DID: <code>{this.props.did}</code></h2>
            <pre><code>{JSON.stringify(this.state.resolutionResult.document, null, 2)}</code></pre>
        </div>
    }
}