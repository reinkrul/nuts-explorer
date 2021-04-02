class DIDList extends React.Component {
    render() {
        return <table>
            <thead>
            <tr>
                <th>DID</th>
                <th>Created</th>
                <th>Updated</th>
            </tr>
            </thead>
            <tbody>
            {this.props.items.map(item =>
                <tr style={{cursor: "pointer"}} onClick={() => this.props.click(item.did)} key={item.did}>
                    <td>{item.did}</td>
                    <td>{item.created}</td>
                    <td>{item.updated}</td>
                </tr>
            )}
            </tbody>
        </table>
    }
}