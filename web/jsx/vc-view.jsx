class VCView extends React.Component {
    state = {
        untrustedVCs: []
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

    render() {
        return <div>
            <VCSearch queryChanged={(concept, query) => this.search(concept, query)}/>
            <h2>Untrusted VCs</h2>
            <VCList items={this.state.untrustedVCs}/>
        </div>;
    }
}