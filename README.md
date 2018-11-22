# go-bitflow
**go-bitflow** is a Go (Golang) library for sending, receiving and transforming streams of data.
The basic data entity is a `bitfolw.Sample`, which consists of a `time.Time` timestamp, a vector of `float64` values, and a `map[string]string` of tags.
Samples can be (un)marshalled in CSV and a dense binary format.
The marshalled data can be transported over files, standard I/O channels, or TCP.
A `SamplePipeline` can be used to pipe a stream of Samples through a chain of transformation or analysis steps implementing the `SampleProcessor` interface.

The `cmd/bitflow-pipeline` sub-package provides an executable with the same name.
The pipeline (including data sources, data sinks, pipeline of transformations steps) is defined by a lightweight, domain-specific scripting language (see subpackage `script`).
Some aspects of the pipeline can also be configured through additional command line flags.

Run `bitflow-pipeline --help` for a list of command line flags.

Go requirement: at least version 1.8.
