package checker

const html = `
	<!DOCTYPE html>
	<htmL lang="en">
	<head>
		<meta charset="utf-8">
		<title>Checker {{ .Name }}</title>
		<style type="text/css">
			html {
				font-size: 16px;
			}
			h1 {
				font-size: 2rem;
			}
			h2 {
				font-size: 1.4rem;
			}
			body {
				background-color: #000511;
			}
			main {
				width: 80%;
				background-color: #001122;
				color: #8888AA;
				border: 1px solid #8888AA;
				border-radius: 25pt;
				margin: 5%;
				padding: 25pt;
				font-family: sans;
				font-weight: 200;
				font-size: 1rem;
			}
			a {
				color: #8888AA;
			}
			.state {
				font-weight: 600;
			}
			.state.OK {
				color: #00AA00;
			}
			.state.Warning  {
				color: #AA5533;
			}
			.state.Failed {
				color: #AA0000;
			}
			dl {
				display: grid;
				grid-template-columns: max-content auto;
			}
			dt {
				grid-column-start: 1;
			}
			dd {
				grid-column-start: 2;
			}
		</style>
	
	</head>
	<body>
		<main>
		<h1>{{ .Name }} <span class="state {{ .State }}">{{ .State }}</span></h1>
		<div>{{ .Summary }}</div>
		<h2>Local Checks</h2>
		<dl>
			{{ range $key, $value := .Checks }}
				<dt><span class="state {{ $value.State }}">{{ $value.State }}</span></dt>
				<dd>{{ $value }}</dd>
			{{ end }}
		</dl>
		<h2>Peers</h2>
		<dl>
			{{ range $key, $value := .PeerStates }}
				<dt><span class="state {{ $value.State }}"> {{ $value.State }}</span></dt>
				<dd>
				<a href="http://{{ $value.Address }}">{{ $value.Name }} ({{ $value.Address }}</a><br/>
				<span class="summary">{{ $value.Summary }}</span>
				<dl>
					{{ range $k, $v := $value.Checks }}
					 	<dt><span class="state {{ $v.State }}">{{ $v.State }}</span></dt>
						<dd>{{ $v }}</dd>
					{{ end }}
				</dl>
				</dd>
			{{ end }}
		</dl>
		</main>
	</body>
	<html>
`
