package feed

import "html/template"

// Tpl is the default template used to generate index.html and page.html files.
// Can be customized using -template command-line argument.
var Tpl = template.Must(template.New("").Parse(`
<!DOCTYPE html><html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<link href="data:image/x-icon;base64,AAABAAEAEBAQAAAAAAAoAQAAFgAAACgAAAAQAAAAIAAAAAEABAAAAAAAgAAAAAAAAAAAAAAAEAAAAAAAAAAhFAUAVUCzAJSHhQD1PycAo6OjAFmWqADo6e0ARCPZAFdUUQDe3t4AL5E6AHvF2wBIgbAAAAAAAAAAAAAAAAAAOqqqaZmZbMyqOqNoSEhszGZmZmZmZmZmiEhIZxEXaEiZmZlnEXdpmYSEhGN3c2SEmZmZYztzaZmISEhjuyNoSJmZmWO7U2mZhISEY4iDZIRmZmZmZmZmZgYAYABgYGAABgBgZgBgBmAGYGAGBgYGAABgYGYGBgYGAGBgBgZmBgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" rel="icon" type="image/x-icon" />
<title>News</title>
<style type="text/css">
html {
	padding: 0;
	margin: 0;
}
body {
	padding: 2em 0;
	margin: 0;
	font-family: sans-serif;
	font-size: 16px;
	background-color:#282828;
}
.container {
	max-width: 1024px;
	margin: 0 auto;
	text-align: center;
}
.item {
	display: block;
	font-size: 18px;
	padding: 1em 0.5em 0.5em 0.5em;
	color: #000;
	font-weight: bold;
	background: #EEE;
	border-bottom: 1px solid #ccc;
}
.tag {
	float: right;
	border: 1px solid #ccc;
	border-radius: 3px;
	padding: 5px;
	margin: -10px 0 0 5px;
    background-color: #ddd;
	font-weight: normal;
}
.item:first-child  {
	border-top: 1px solid #aaa;
}
.item:visited {
	color: #873B3B;
}
.item, .item:visited {
	text-decoration: none;
	text-align: left;
}
.item:hover, .item:active {
	text-decoration: underline;
}
.next {
    display: inline-block;
    margin: 20px auto;
    font-size: 24px;
    padding: 0.5em;
    color: #000;
    font-weight: bold;
    background: #EEE;
    border-bottom: 1px solid #ccc;
    border-radius: 5px;
    text-decoration: none;
}
.next:hover, .next:active {
	text-decoration: underline;
}
.feed {
	display: none;
}
</style>
</head>
<body>
{{range $url, $title := .Feeds}}
<a class="feed" href="{{$url}}">{{$title}}</a>{{end}}
<div class="container">{{range .Items}}
<a class="item" target="_blank" href="{{.URL}}">{{if .Tag}}<div class="tag">{{.Tag}}</div>{{end}}{{.Title}}</a>{{end}}
{{if gt .NextPage 1}}<a class="next" href="page{{.NextPage}}.html">Next</a>{{end}}
</div>

</body>
</html>`))
