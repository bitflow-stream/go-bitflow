# Plugin can later be loaded in pipeline (optional parameters at the end of the URL):
# plugin://bitflow-plugin-mock/Plugin?interval=200ms&offset=1h&error=10
go build -buildmode=plugin -o bitflow-plugin-mock .
