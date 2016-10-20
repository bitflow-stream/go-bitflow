# go-bitflow-pipeline
go-bitflow-pipeline is a Go (Golang) tool for sending, receiving and transforming streams of data.
It uses the `github.com/antongulenko/data2go` library for sending and receiving instances of `Sample`, and adds many implementations of `SampleProcessor` that can be chained together to form an `AlgorithmPipeline`.
The `bitflow-pipeline` sub-package provides an executable with the same name.
The data sources, data sinks, pipeline of transformations steps, and other parameters can be configured through numerous command line flags.

Run `bitflow-pipeline --help` for a list of command line flags.

## Installation:
* Install git and go (at least version **1.6**).
* Make sure `$GOPATH` is set to some existing directory.
* Execute the following command to make `go get` work with Gitlab:

```shell
git config --global "url.git@gitlab.tubit.tu-berlin.de:CIT-Huawei/go-bitflow-pipeline.git.insteadOf" "https://github.com/antongulenko/go-bitflow-pipeline"
```
* Get and install this tool:

```shell
go get github.com/antongulenko/go-bitflow-pipeline/bitflow-pipeline
```
* The binary executable `bitflow-pipeline` will be compiled to `$GOPATH/bin`.
 * Add that directory to your `$PATH`, or copy the executable to a different location.

