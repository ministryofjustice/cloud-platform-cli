## cloud-platform completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(cloud-platform completion bash)

To load completions for every new session, execute once:

#### Linux:

	cloud-platform completion bash > /etc/bash_completion.d/cloud-platform

#### macOS:

	cloud-platform completion bash > $(brew --prefix)/etc/bash_completion.d/cloud-platform

You will need to start a new shell for this setup to take effect.


```
cloud-platform completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform completion](cloud-platform_completion.md)	 - Generate the autocompletion script for the specified shell

