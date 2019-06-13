# ioFog Unified CLI

`iofogctl` is a CLI for the installation, configuration, and operation of ioFog Edge Compute Networks (ECNs).
It can be used to remotely manage multiple different clusters from a single host. It is built for an
ioFog user and a DevOps engineering wanting to manage ioFog clusters.  

## Prerequisites

The following must be installed and configured before performing bootstrap:
* Go 1.12.1+

## Install

Mac users can use Homebrew:

```bash
brew tap eclipse-iofog/iofogctl
brew install iofogctl
```

Otherwise, `iofogctl` can be installed in the usual Go fashion:

```bash
go get -u github.com/eclipse-iofog/iofogctl/cmd/ifogctl
```

Install dependencies:

```
script/bootstrap.sh
```
## Usage

#### Quick Start

See all iofogctl options

```
iofogctl --help
```

Current options include:

```
ioFog Unified Command Line Interface

Usage:
  iofogctl [command]

Available Commands:
  create      Create an ioFog resource
  delete      Delete existing ioFog resources
  deploy      Deploy ioFog stack on existing infrastructure
  describe    Get detailed information of existing resources
  get         Get information of existing resources
  help        Help about any command
  legacy      Execute commands using legacy CLI
  logs        Get log contents of deployed resource

Flags:
      --config string      CLI configuration file (default is ~/.iofog.yaml)
  -h, --help               help for iofogctl
  -n, --namespace string   Namespace to execute respective command within (default "default")

Use "iofogctl [command] --help" for more information about a command.
```

## Building 

If you want to build from the src, you can see all `make` commands by running:
```
make help
```

Easy path is to build and install:
```
make all
iofogctl --help
```

iofogctl is installed in your `$GOPATH/bin`

## Running Tests

Run project unit tests:

```
make test
```

This will output an JUnit compatible 'test-report.xml' file that can be imported in most CI systems.