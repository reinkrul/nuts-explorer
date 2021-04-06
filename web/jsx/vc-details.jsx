class VCDetails extends React.Component {
    render() {
        return <div>
            <h2>VC: <code>{this.props.document.verifiableCredential.id}</code></h2>
            <pre><code>{JSON.stringify(this.props.document, null, 2)}</code></pre>
        </div>
    }
}