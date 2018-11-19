# go-bitflow-pipeline
go-bitflow-pipeline is a Go (Golang) tool for sending, receiving and transforming streams of data.
It uses the `github.com/bitflow-stream/go-bitflow` library for sending and receiving instances of `Sample`, and adds many implementations of `SampleProcessor` that can be chained together to form a `SamplePipeline`.
The `bitflow-pipeline` sub-package provides an executable with the same name.
The pipeline (including data sources, data sinks, pipeline of transformations steps) is defined by a lightweight, domain-specific scripting language (see subpackage `query`).
Some aspects of the pipeline can also be configured through additional command line flags.

Run `bitflow-pipeline --help` for a list of command line flags.

Go requirement: at least version 1.8.
