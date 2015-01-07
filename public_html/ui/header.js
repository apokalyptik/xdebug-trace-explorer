var Header = React.createClass({
	getInitialState: function() {
		return {"data": {}}
	},
	componentDidMount: function() {
		$.ajax({
			url: "/api/info.json",
			dataType: 'json',
			success: function(data) {
				this.setState({"data": data});
			}.bind(this)
		});
	},
	render: function() {
		return( <h1>{this.state.data.filename} <em>{this.state.data.filesize}</em></h1> );
	}
})
