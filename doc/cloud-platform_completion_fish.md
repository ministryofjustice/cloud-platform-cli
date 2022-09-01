## cloud-platform completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	cloud-platform completion fish | source

To load completions for every new session, execute once:

	cloud-platform completion fish > ~/.config/fish/completions/cloud-platform.fish

You will need to start a new shell for this setup to take effect.


```
cloud-platform completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform completion](cloud-platform_completion.md)	 - Generate the autocompletion script for the specified shell

