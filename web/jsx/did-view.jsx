class DIDView extends React.Component {
    state = {
        dids: [],
        currentDID: null,
    };

    componentDidMount() {
        new DIDService('.').list().then((list) => {
            this.setState({dids: list});
        })
    }

    render() {
        return <div>
            <DIDList items={this.state.dids} click={(did) => this.setState({currentDID: did})}/>
            {this.state.currentDID ? <DIDDetails did={this.state.currentDID}/> : ""}
        </div>;
    }
}