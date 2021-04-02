class VCView extends React.Component {
    state = {
        untrustedVCs: [],
        currentDetailsVCID: null,
    };

    componentDidMount() {

    }

    search(concept, query) {
        if (Object.entries(query).length > 0) {
            new VCService().search(concept, query).then((items) => {
                console.log(items);
            })
        }
    }

    showDetails(id) {
        new VCService().get(id).then((doc) => this.setState({currentDetailsVCID: id}));
    }


    render() {
        return <div>
            <VCSearch queryChanged={(concept, query) => this.search(concept, query)}/>
            {this.state.currentDetailsVCID ? <VCDetails id={this.state.currentDetailsVCID}/> : ""}
            <h2>Untrusted VCs</h2>
            <VCList items={this.state.untrustedVCs} click={(id) => this.showDetails(id)}/>
        </div>;
    }
}