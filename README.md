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

### Documentation

The entire CLI documentation can be found [here](https://github.com/eclipse-iofog/iofogctl/blob/develop/docs/md/iofogctl.md)

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
      --config string      CLI configuration file (default is ~/.iofog/config.yaml)
  -h, --help               help for iofogctl
  -n, --namespace string   Namespace to execute respective command within (default "default")

Use "iofogctl [command] --help" for more information about a command.
```

### Autocomplete

If you are running BASH or ZSH, iofogctl comes with shell autocompletion scripts.
In order to generate those scripts, run:

```bash
iofogctl autocomplete bash
```
OR

```bash
iofogctl autocomplete zsh
```

Then follow the instructions output by the command.

Example:
```bash
$> iofogctl autocomplete bash
✔ $HOME/.iofog/completion.bash.sh generated
Run `source $HOME/.iofog/completion.bash.sh` to update your current session
Add `source $HOME/.iofog/completion.bash.sh` to your bash profile to have it saved

$>source $HOME/.iofog/completion.bash.sh
$>echo "$HOME/.iofog/completion.bash.sh" >> $HOME/.bash_profile
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
