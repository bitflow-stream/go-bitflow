
# A plugin can be loaded using the -p switch of bitflow-pipeline:
#  bitflow-pipeline -p bitflow-plugin-mock
# This plugin will load the following:
#  1. A data source, that can be used the following way in a bitflow script. It will produce random data in the given intervals and produce an error after a fixed number of samples.
#     mock://interval=200ms&offset=1h&error=10 -> ...
#  2. A data processor, used like this. It will print incoming samples in the given frequency and produce an error after a fixed number of samples.
#     ... -> mock(print=10, error=10) -> ...

# Install the pipeline to make sure the plugin is built against an up-to-date binary
echo "Building go-bitflow-pipeline..."
go install github.com/bitflow-stream/go-bitflow-pipeline/...

# Now build the plugin (use the name of the folder as binary target)
echo "Building bitflow-plugin-mock..."
home=`dirname $(readlink -e $0)`
plugin_name=$(basename "$home")
cd "$home" && go build -buildmode=plugin -o "$plugin_name" .
