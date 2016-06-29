# data2go
Several Go (Golang) packages for collecting and working with streams of data

## Installation:
* Install git, go (at least version **1.6**) and the package **libvirt-dev**.
* Make sure `$GOPATH` is set to some existing directory
* Execute the following two commands to make `go get` work with gitlab:

```shell
git config --global "url.git@gitlab.tubit.tu-berlin.de:CIT-Huawei/monitoring.git.insteadOf" "https://github.com/antongulenko/data2go"
git config --global "url.git@gitlab.tubit.tu-berlin.de:.insteadOf" "https://gitlab.tubit.tu-berlin.de/"
```
* Get and install this repository and all sub-packages:
 * `go get github.com/antongulenko/data2go/...`
* The programs `data-collection-agent` and `analysis-pipeline` are now in `$GOPATH/bin`. Add `$GOPATH/bin` to your `$PATH` to use them directly.