## cloud-platform pipeline cordon-and-drain

cordon and drain a node group on a cluster

### Synopsis


Running this command will cordon and drain an existing node group in a eks cluster in the cloud-platform aws account.
   It will not terminate the nodes nor will it delete the node group.

   The cordon and drain will run remotely in the pipeline, and under the hood it calls "cloud-platform cluster recycle-node --name <node-name> --drain-only --ignore-label"

You must have the following environment variables set, or passed via arguments:
	- a cluster name

   ** You _must_ have the fly cli installed **
   --> https://concourse-ci.org/fly.html

   ** You must also have wget installed **
   --> brew install wget


```
cloud-platform pipeline cordon-and-drain [flags]
```

### Options

```
      --cluster-name string   cluster to run pipeline cmds against
  -h, --help                  help for cordon-and-drain
      --node-group string     node group name to cordon and drain
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform pipeline](cloud-platform_pipeline.md)	 - Cloud Platform pipeline actions

