const Views = [
    {
        title: "Decentralized Identifiers",
        render: () => <DIDView/>,
    },
    {
        title: "Verifiable Credentials",
        render: () => <VCView/>,
    },
    {
        title: "Network Viewer",
        render: () => <NetworkView/>,
    },
    {
        title: "DAG Viewer",
        render: () => <DAGView/>,
    },
]


class Sidebar extends React.Component {
    render() {
        return <div className="bg-dark p-3" style={{height: "100vh"}}>
            <a href="/" className="text-white">
                <h2 className="fs-4">Nuts Explorer</h2>
            </a>
            <hr/>
            <ul className="nav nav-pills flex-column mb-auto">
                {Views.map(view => {
                    let classes = ["nav-link"]
                    if (this.props.activeView === view.title) {
                        classes.push("active")
                    } else {
                        classes.push("text-white")
                    }
                    return <li key={"viewlink-"+view.title}><a href="#" className={classes.join(" ")} onClick={() => this.props.navigateView(view.title)}>{view.title}</a></li>
                })}
            </ul>
        </div>;
    }
}

class Application extends React.Component {
    state = {
        activeView: Views[0].title
    }

    render() {
        const view = Views.filter((view) => this.state.activeView === view.title)[0]
        return <div className="container-fluid">
            <div className="row">
                <div className="col-md-auto px-0">
                    <Sidebar activeView={this.state.activeView} navigateView={(newView) => this.setState({activeView: newView})} />
                </div>
                <div className="col py-3">
                    <h1>{view.title}</h1>
                    {view.render()}
                </div>
            </div>
        </div>;
    }
}