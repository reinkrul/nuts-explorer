class VCSearch extends React.Component {
    state = {
        concept: "organization",
        query: {},
    };

    updateQuery(key, value) {
        let newQuery = this.state.query
        if (value === "") {
            delete newQuery[key]
        } else {
            newQuery[key] = value
        }
        this.setState({query: newQuery})
        this.props.queryChanged(this.state.concept, this.state.query)
    }

    render() {
        return <form>
            <h3>Search for NutsOrganizationCredential</h3>
            <div className="form-group row">
                <label htmlFor="searchVCName" className="col-sm-2 col-form-label">Name:</label>
                <div className="col-sm-10">
                    <input type="text" id="searchVCName" onKeyUp={(e) => this.updateQuery('name', e.target.value)}/>
                </div>
            </div>
            <div className="form-group row">
                <label htmlFor="searchVCCity" className="col-sm-2 col-form-label">City:</label>
                <div className="col-sm-10">
                    <input type="text" id="searchVCCity" onKeyUp={(e) => this.updateQuery('city', e.target.value)}/>
                </div>
            </div>
            <code>{JSON.stringify(this.state.query)}</code>
        </form>
    }
}