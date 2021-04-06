class VCView extends React.Component {
    constructor(props) {
        super(props)
        this.refresh()
    }

    state = {
        untrustedIssuers: [],
        trustedIssuers: [],
        searchResults: [],
        currentDetailsVC: null,
    };

    refresh() {
        new VCService().getUntrustedIssuers().then((items) => {
            this.setState({untrustedIssuers: items})
        })
        new VCService().getTrustedIssuers().then((items) => {
            this.setState({trustedIssuers: items})
        })
    }

    search(concept, query) {
        if (Object.entries(query).length > 0) {
            new VCService().search(concept, query).then((items) => {
                this.setState({searchResults: items})
            })
        }
    }

    showDetails(vc) {
        new VCService().get(vc.id).then((doc) => this.setState({currentDetailsVC: doc}));
    }

    render() {
        return <div>
            <VCSearch queryChanged={(concept, query) => this.search(concept, query)}/>
            <VCList items={this.state.searchResults} click={(id) => this.showDetails(id)}/>
            {this.state.currentDetailsVC ? <VCDetails document={this.state.currentDetailsVC}/> : ""}
            <h2>Trusted VC Issuers</h2>
            <VCIssuers items={this.state.trustedIssuers}/>
            <h2>Untrusted VC Issuers</h2>
            <VCIssuers items={this.state.untrustedIssuers}/>
        </div>;
    }
}