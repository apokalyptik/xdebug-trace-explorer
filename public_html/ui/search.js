var Search = React.createClass({
	getInitialState: function() {
		return { data: { open: false, success: false, results: {} }};
	},
	style: function(type) {
		if ( type == "div" ) {
			return {
				position: "absolute",
				"top": "0.5em",
				right: "0.5em",
			}
		}
		if ( type == "input" ) {
			return {
				fontSize: "1.5em",
				width: "15em"
			}
		}
	},
	render: function() {
		return (
			<div style={this.style("div")}>
				<input style={this.style("input")} name="search" placeholder="function search"/>
			</div>
		);
	}
});
