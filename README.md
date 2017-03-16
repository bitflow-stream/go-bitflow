# go-bitflow-pipeline
go-bitflow-pipeline is a Go (Golang) tool for sending, receiving and transforming streams of data.
It uses the `github.com/antongulenko/go-bitflow` library for sending and receiving instances of `Sample`, and adds many implementations of `SampleProcessor` that can be chained together to form an `SamplePipeline`.
The `bitflow-pipeline` sub-package provides an executable with the same name.
The pipeline (including data sources, data sinks, pipeline of transformations steps) is defined by a lightweight, domain-specific scripting language (see below).
Some aspects of the pipeline can also be configured through additional command line flags.

Run `bitflow-pipeline --help` for a list of command line flags.

## Installation:
* Install git and go (at least version **1.6**).
* Make sure `$GOPATH` is set to some existing directory.
* Execute the following command to make `go get` work with Gitlab. This requires a passwordless SSH connection to the Gitlab server.

```shell
git config --global "url.git@gitlab.tubit.tu-berlin.de:anton.gulenko/go-bitflow-pipeline.git.insteadOf" "https://github.com/antongulenko/go-bitflow-pipeline"
```

* If the passwordless SSH connection does not work, you can manually clone the repository via HTTPS:

```shell
mkdir -p "$GOPATH/src/github.com/antongulenko"
git clone https://gitlab.tubit.tu-berlin.de/CIT-Huawei/go-bitflow-pipeline.git "$GOPATH/src/github.com/antongulenko/go-bitflow-pipeline" 
```

* Get and install this tool:

```shell
go get github.com/antongulenko/go-bitflow-pipeline/bitflow-pipeline
```

* The binary executable `bitflow-pipeline` will be compiled to `$GOPATH/bin`.
 * Add that directory to your `$PATH`, or copy the executable to a different location.

## TODO
* Document the bitflow script language
