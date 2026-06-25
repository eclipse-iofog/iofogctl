## potctl completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(potctl completion zsh)

To load completions for every new session, execute once:

#### Linux:

	potctl completion zsh > "${fpath[1]}/_potctl"

#### macOS:

	potctl completion zsh > $(brew --prefix)/share/zsh/site-functions/_potctl

You will need to start a new shell for this setup to take effect.


```
potctl completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
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


