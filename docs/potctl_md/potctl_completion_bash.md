## potctl completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(potctl completion bash)

To load completions for every new session, execute once:

#### Linux:

	potctl completion bash > /etc/bash_completion.d/potctl

#### macOS:

	potctl completion bash > $(brew --prefix)/etc/bash_completion.d/potctl

You will need to start a new shell for this setup to take effect.


```
potctl completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl completion](potctl_completion.md)	 - Generate the autocompletion script for the specified shell


