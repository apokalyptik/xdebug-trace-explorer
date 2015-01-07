var FunctionView = React.createClass({
	getInitialState: function() {
		window.onhashchange = this.componentDidMount.bind(this)
		return { data: {
			id: 0,
			parent_id: 0,
			details: {
				func: "",
				args: []
			},
			children: [],
		} }
	},
	switchFunc: function(n) {
		$.ajax({
			url: "/api/func.json?n=" + n,
			dataType: 'json',
			success: function(data) {
				if ( data.data.child_count > 0 ) {
					var maxt = 0
					data.data.children.sort(function(a,b) {
						return a.id - b.id
					})
				}
				window.location = "#n=" + n
				this.setState(data);
			}.bind(this)
		})
	},
	componentDidMount: function() {
		var n = 0
		if ( window.location.hash.substr(1) != "" ) {
			n = window.location.hash.substr(3)
		}
		$.ajax({
			url: "/api/func.json?n=" + n,
			dataType: 'json',
			success: function(data) {
				data.data.children.sort(function(a,b) {
					return a.id - b.id
				})
				this.setState(data);
			}.bind(this)
		})
	},
	render: function() {
		var Children = this.state.data.children.map(function(c, ci) {
			var onClick = function(event) {
				event.stopPropagation();
			}
			var numArgs = c.details.args.length - 1
			var ArgDots = c.details.args.map(function(a,ai) {
				var text = ""
				var link = true
				if ( a.length < 50 ) {
					text=a
					link = false
				} else {
					text=a.substring(0,47) + "..."
				}
				if ( numArgs > 0 ) {
						if ( ai < numArgs ) {
							text = text + ", "
						} else {
							text = text
						}
				} else {
					text = text
				}
				if ( link ) {
					LinkHref = "#n=" + c.parent_id
					return (
						<span>
						<a href={LinkHref} onClick={function(event) {
							$.modal("<textarea class='rawdata'>"+c.details.args[0]+"</textarea>")
							event.stopPropagation();
						}.bind(c)}>{text}</a>
						</span>
					);
				}
				return (<span>{text}</span>);
			})
			var MoreLink = ""
			var MoreClick = function(event) {
				this.switchFunc(c.id);
				event.stopPropagation();
			}.bind(this)

			if ( c.child_count > 0 ) {
				console.log(c)
				MoreLink = ( 
					<div className="fndive">
						<a href={window.location.href} onClick={MoreClick}>&gt;&gt;</a>
					</div>
				);
			}


			var pct = 0
			if ( this.state.data.duration > 0 ) {
				pct = ( ( c.duration / this.state.data.duration ) * 100 )
			}
			var hot="hot1"
			if ( pct > 50 ) {
				hot="hot4"
			} else if ( pct > 25 ) {
				hot="hot3"
			} else if ( pct > 10 ) {
				hot="hot2"
			}
			hot = "fnpre " + hot

			var RawID = "raw-" + c.id
			var Raw = c.raw
			var ShowRaw = function(event) {
				$.modal("<textarea class='rawdata'>"+c.raw+"</textarea>")
				event.stopPropagation()
			}.bind(c)
			var LinkHref = "#n=" + c.parent_id
			return (
			<div>
				<div className="child">
					<div style={{padding: "0.5em"}} className={hot}><a href={LinkHref} onClick={ShowRaw}>#{c.id}</a></div>
					<div className="fnpre">
						Mem: {c.memory}<br/>
						WCT: {c.duration}
					</div>
					<div className="fn">
						{c.details.func}({ArgDots})<br/>
						{c.location.file}@{c.location.line}
					</div>
					{MoreLink}
				</div>
			</div>
			);
		}.bind(this))

		var BackLink = ""
		var ID = this.state.data.id
		var PID = this.state.data.parent_id
		if ( this.state.data.id != this.state.data.parent_id ) {
			var BackClick = function(event) {
				this.switchFunc(PID);
				event.stopPropagation();
			}.bind(this)
			BackLink = (<div className="fnjump"><a href={window.location.href} onClick={BackClick}>&lt;&lt;</a></div>);
		}
		var locationInfo = ""
		if ( 'undefined' != typeof this.state.data.location && this.state.data.location.line ) {
			locationInfo = ( <span>{this.state.data.location.file}@{this.state.data.location.line}</span> )
		}
		return( 
			<div>
				<div id="funchead">
					{BackLink} <span className="fnhdeets">
						{this.state.data.details.func}() {locationInfo}
					</span>
				</div>
				<div id="childlist">
					{Children}
				</div>
			</div>
		);
	}
});
