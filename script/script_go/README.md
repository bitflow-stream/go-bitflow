# Bitflow Script

A Bitflow Script mostly ignores white space (spaces, tabs and newlines). There are no comments.

The primary element of a Bitflow Script is a *processing step*.
A processing step is defined through an identifier, followed by parameters in parentheses.
The parameters are key-value string pairs, separated through `=` und `,` characters:
```
avg()
```
```
tag(key=value)
```
```
store_pca_model(precision = 0.99, file = pca-model.bin)
```

Identifiers and strings are the same in Bitflow Script.
They can both be written without quotes, but are interrupted by white space and special characters (``( ) , = ; -> { } [ ] ' " ` ``).
Strings can be quotes with pairs of the following characters: ``'  " ` ``.
There are three different quotes to assure flexibility when embedding different kinds of strings into each other.
The parameters and names of processing steps can also be written as quoted strings:
```
"avg"()
```
```
'avg'()
```
```
`my "special" avg`()
```
```
'store_pca_model'("precision"=`0.99`, `file`="pca-model.bin")
```

Processing steps can be chained together to form a pipeline. This is noted through the `->` character sequence:
```
inputfile.csv -> filter(host=mypc) -> avg() -> 192.168.0.55:5555
```

The first and last step of a pipeline can optionally be a data source or a data sink, respectively.
Both a source and a sink are optional, and only allowed at the first and last position.
A source or sink is provided by defining a string without parameters (e.g. `/home/inputfile.csv`, `tcp://localhost:5555`).
A source or sink with parameters can be defined through a quoted string (e.g. `'/home/data (final).csv'`).
In a pipeline that contains *only* a single data sink or source definition, it is interpreted as a data *source*.

Multiple pipelines can be executed in parallel. This is done by separating them through ';' characters and surrounding them by curly braces:
```
{
	input1.csv -> avg();
	input2.csv -> filter(host=mypc) -> 192.168.0.55:5555
}
```

Parallel pipelines can also be used as processing step and arbitrarily nested:
```
input.csv -> { avg() ; max() } -> output.csv
```
```
{
	input.csv -> { avg(); max() } -> 192.168.0.1:5555;
	:6666 -> output.csv
}
```

In the case of using parallel pipelines as processing step, every sample that reaches the opening curly brace will be piped through all defined sub pipelines in parallel.
The step right after the closing curly brace will receive the results of all the sub pipelines.

The simple curly braces are an abbreviated notation of the multiplex fork processing step:
```
input.csv -> multiplex(num=2) { avg() ; max() } -> output.csv
```

This notation of combining a processing step with parameters, followed by parallel pipelines, is called a fork processing steps.
There are two basic forks, but further forks can be programed for use case specific tasks.
* rr: Round Robin fork, sends the incoming samples through the defined pipelines in a Round Robin fashion
* mulitplex: Sends a copy of each incoming Sample into every sub pipeline

### Definition of data sources and sinks

A data source or sink (called data *store* here) is denoted by a simple string without parameters.
The four main data stores are files, TCP listen sockets, active TCP connections, and the standard in/out streams.
The kind of source is detected automatically from the string:
- A *host:port* pair like `192.160.0.55:4444` is interpreted as an active TCP connection, meaning a connection to that endpoint will be established.
  Depending on whether this data store is used as a data source or data sink, the connection will be used to either send or receive data.
- A separate *:port* like `:5555` is interpreted as a TCP listening socket. A socket will be opened, and waitfor incoming connections.
  Sockets can also be used to both receive and serve data.
- A single hyphen `-` is interpreted as standard input or output.
- Any other string is interpreted as a file for writing or reading.

If the automatic data store type is not desired, it can be explicitly specified with a URL-like notation:
* `tcp://xxx` will interpret the rest of the string as a TCP endpoint to actively connect to (can also be a single port without hostname, in which case it will connect to localhost)
* `listen://xxx` will interpret the string as a local TCP endpoint to listen on (can include an IP address or hostname for disambiguation)
* `file://xxx` will force read or write to a file
* `empty://xxx` will produce no input or output, the rest of the string is ignored (useful in rare cases)
* `std://-` will read or write from standard input or output, only valid with the `-` string
* `box://-` is a special transport type only usable as output, which displays the data in a continuously updated table on the console

Aside of the transport type, the format of the data is important.
The data can be transported and stored in two main formats: CSV (`csv`) and binary (`bin`).
A third data format is the `text` format, which can only be used as output and is helpful to display data in a human-readable way on the console.
When receiving data, the format is usually detected automatically.
For writing data, there are default data formats:
* `csv` is used for `file`
* `bin` is used for `tcp` and `listen`, and if a `file` has a `.bin` suffix
* `text` is used for `std`
* `box` and `empty` do not depend on the data format

The format can also be forced for input data by specifying it in the same way.

If both a data format and a data store type are specified, they are separated through `+` signs:
```
listen+csv://10.0.0.1:4444
```

##### TODO Missing explanations
* Brackets
* Special case of parallel pipelines in the beginning of the top-level pipeline
