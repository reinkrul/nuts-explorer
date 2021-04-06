class VCList extends React.Component {
   render() {
        return <table>
            <thead>
            <tr>
                <th>ID</th>
                <th>Issuer</th>
                <th>Subject</th>
            </tr>
            </thead>
            <tbody>
            {this.props.items.map(item =>
                <tr style={{cursor: "pointer"}} onClick={() => this.props.click(item)} key={item.id}>
                    <td>{item.id}</td>
                    <td>{item.issuer}</td>
                    <td>{item.subject}</td>
                </tr>
            )}
            </tbody>
        </table>
    }
}