[![Build Status](https://ci.bitflow.team/jenkins/buildStatus/icon?job=Bitflow%2Fgo-bitflow%2Fmaster&build=lastBuild)](http://wally144.cit.tu-berlin.de/jenkins/blue/organizations/jenkins/Bitflow%2Fgo-bitflow/activity)
[![Code Coverage](https://ci.bitflow.team/sonarqube/api/project_badges/measure?project=go-bitflow&metric=coverage)](http://wally144.cit.tu-berlin.de/sonarqube/dashboard?id=go-bitflow)
[![Maintainability](https://ci.bitflow.team/sonarqube/api/project_badges/measure?project=go-bitflow&metric=sqale_rating)](http://wally144.cit.tu-berlin.de/sonarqube/dashboard?id=go-bitflow)
[![Reliability](https://ci.bitflow.team/sonarqube/api/project_badges/measure?project=go-bitflow&metric=reliability_rating)](http://wally144.cit.tu-berlin.de/sonarqube/dashboard?id=go-bitflow)

# go-bitflow
**go-bitflow** is a Go (Golang) library for sending, receiving and transforming streams of data.
`go-bitflow` is mainly used in a Kubernetes cluster, managed by the [Bitflow K8s Operator](https://github.com/bitflow-stream/bitflow-k8s-operator), but can also be executed standalone either in a Docker container or as a native binary.

The basic data entity is a [`bitfolw.Sample`](bitflow/sample.go), which consists of a `time.Time` timestamp, a vector of `float64` values, and a `map[string]string` of tags.
Samples can be (un)marshalled in CSV and a dense binary format.
The marshalled data can be transported over files, standard I/O channels, or TCP.
A [`bitflow.SamplePipeline`](bitflow/pipeline.go) can be used to pipe a stream of Samples through a tree of transformation or analysis steps implementing the [`bitflow.SampleProcessor`](bitflow/sample_processor.go) interface.
The [`cmd/bitflow-pipeline`](cmd/bitflow-pipeline) sub-package provides an executable with the same name.
The dataflow graph (including data sources, data sinks, tree of operators) is expressed in the simple domain specific language [`Bitflowscript`](https://bitflow.readthedocs.io/projects/bitflow-antlr-grammars/en/latest/bitflow-script).
Some aspects of the execution can also be configured through additional command line flags.

Go requirement: at least version 1.11. Use the `--help` flag for a list of command line flags.

# Installation

### Native installation

```
go install github.com/bitflow-stream/go-bitflow/cmd/bitflow-pipeline
```

### Build Docker container

There are Dockerfiles for different processors: `AMD64` (using Alpine Linux), `arm32v7` and `arm64v8`.
In addition, there are Dockerfiles for "static" builds, that produce a container from `scratch`, that contains only the executable `bitflow-pipeline` file.
All containers can be either built fully with one Dockerfile (clean but slow), or use a local build and then simply copy the resulting binary in the container (utilized Go mod and build cache).

All commands below are executed in the repository root.

#### Full build (slow)

Select one of the files in `build/multi-stage/[alpine|arm23v7|arm64v8]-[full|static].Dockerfile` and run:

```
docker build -t [IMAGE_NAME] -f build/multi-stage/[...].Dockerfile
```

#### Cached static build (fast & small container)

```
./build/native-static-build.sh
docker build -t [IMAGE_NAME] -f build/static-prebuilt.Dockerfile build/_output/static
```

#### Cached build (ARM or Alpine)

Select a build target `arm32v7|arm64v8|alpine` and a directory to store the Go build cache (`/tmp/go-cache-directory` in this example).

```
TARGET=[arm32v7|arm64v8|alpine]
./build/containerized-build.sh $TARGET /tmp/go-cache-directory
docker build -t [IMAGE_NAME] -f build/$TARGET-prebuilt.Dockerfile build/_output/$TARGET
```
