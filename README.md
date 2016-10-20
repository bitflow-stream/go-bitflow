# data2go
data2go is a Go (Golang) library for sending, receiving and transforming streams of data.
The basic data entity is a `Sample`, which consists of a `time.Time` timestamp, a vector of `float64` values, and a `map[string]string` of tags.
Samples can be (un)marshalled in CSV and a dense binary format.
The marshalled data can be transported over files, standard I/O channels, or TCP.
A `SamplePipeline` can be used to pipe a stream of Samples through a chain of transformation or analysis steps implementing the `SampleProcessor` interface.

## Installation:
* Install git and go (at least version **1.6**).
* Make sure `$GOPATH` is set to some existing directory.
* Execute the following command to make `go get` work with Gitlab:

```shell
git config --global "url.git@gitlab.tubit.tu-berlin.de:CIT-Huawei/data2go.git.insteadOf" "https://github.com/antongulenko/data2go"
```
* Get and install this library:

```shell
go get github.com/antongulenko/data2go
```
