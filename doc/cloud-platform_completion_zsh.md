## cloud-platform completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(cloud-platform completion zsh); compdef _cloud-platform cloud-platform

To load completions for every new session, execute once:

#### Linux:

	cloud-platform completion zsh > "${fpath[1]}/_cloud-platform"

#### macOS:

	cloud-platform completion zsh > $(brew --prefix)/share/zsh/site-functions/_cloud-platform

You will need to start a new shell for this setup to take effect.


```
cloud-platform completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform completion](cloud-platform_completion.md)	 - Generate the autocompletion script for the specified shell

