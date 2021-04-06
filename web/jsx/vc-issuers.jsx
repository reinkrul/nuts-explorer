class VCIssuers extends React.Component {
    render() {
        const items = Object.entries(this.props.items).map(entry => {
            return entry[1].map(did => {
                return {type: entry[0], did: did}
            })
        }).flat()
        return <table>
            <thead>
            <tr>
                <th>Type</th>
                <th>Issuer</th>
            </tr>
            </thead>
            <tbody>
            {items.map(item =>
                <tr key={"vc-issuer-" + item.type + item.did}>
                    <td>{item.type}</td>
                    <td>{item.did}</td>
                </tr>
            )}
            </tbody>
        </table>
    }
}