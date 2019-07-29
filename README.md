# ioFog Unified CLI

`iofogctl` is a CLI for the installation, configuration, and operation of ioFog Edge Compute Networks (ECNs).
It can be used to remotely manage multiple different clusters from a single host. It is built for an
ioFog user and a DevOps engineering wanting to manage ioFog clusters.  

## Install

#### Mac

Mac users can use Homebrew:

```bash
brew tap eclipse-iofog/iofogctl
brew install iofogctl
```

#### Linux

The Debian package can be installed like so:
```bash
https://packagecloud.io/install/repositories/iofog/iofogctl/script.deb.sh | sudo bash
sudo apt install iofogctl
```

And similarly, the RPM package can be installed like so:
```
https://packagecloud.io/install/repositories/iofog/iofogctl/script.rpm.sh | sudo bash
sudo apt install iofogctl
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

## Build from Source

Go 1.12.1+ is a prerequisite. Install all other dependancies with:
```
script/bootstrap.sh
```

See all `make` commands by running:
```
make help
```

To build and install, go ahead and run:
```
make all
iofogctl --help
```

iofogctl is installed in `/usr/local/bin`

## Running Tests

Run project unit tests:
```
make test
```

This will output a JUnit compatible file into `reports/TEST-iofogctl.xml` that can be imported in most CI systems.
