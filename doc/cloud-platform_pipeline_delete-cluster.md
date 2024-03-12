## cloud-platform pipeline delete-cluster

delete a cloud-platform cluster via the pipeline

### Synopsis


Running this command will delete an existing eks cluster in the cloud-platform aws account.
It will delete the components, then the cluster, the VPC, terraform workspace and cleanup the cloudwatch log group.

   The delete will run remotely in the pipeline

You must have the following environment variables set, or passed via arguments:
	- a cluster name

Optionally you can pass a branch name to use for the pipeline run, default is "main"

   ** You _must_ have the fly cli installed **
   --> https://concourse-ci.org/fly.html

   ** You must also have wget installed **
   --> brew install wget


```
cloud-platform pipeline delete-cluster [flags]
```

### Options

```
  -b, --branch-name string    branch name to use for pipeline run (default: main) (default "main")
      --cluster-name string   cluster to delete
  -h, --help                  help for delete-cluster
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform pipeline](cloud-platform_pipeline.md)	 - Cloud Platform pipeline actions

