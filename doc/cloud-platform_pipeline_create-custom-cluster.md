## cloud-platform pipeline create-custom-cluster

create a custom cloud-platform cluster via the pipeline

### Synopsis


Running this command will create a custom eks cluster in the cloud-platform aws account.

   This command will run remotely in the pipeline

Optionally you can pass a branch name to use for the pipeline run, default is "main"

   ** You _must_ have the fly cli installed **
   --> https://concourse-ci.org/fly.html

   ** You must also have wget installed **
   --> brew install wget


```
cloud-platform pipeline create-custom-cluster [flags]
```

### Options

```
  -b, --branch-name string   branch name to use for pipeline run (default: main) (default "main")
  -h, --help                 help for create-custom-cluster
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform pipeline](cloud-platform_pipeline.md)	 - Cloud Platform pipeline actions

