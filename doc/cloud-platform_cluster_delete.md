## cloud-platform cluster delete

delete a cloud-platform cluster

### Synopsis


Running this command will delete an existing eks cluster in the cloud-platform aws account.
It will delete the components, then the cluster, the VPC, terraform workspace and cleanup the cloudwatch log group.

This command defaults to --dry-run=true, set --dry-run=false when you are ready to actually destroy the cluster.

You must have the following environment variables set, or passed via arguments:
	- a valid AWS profile or access key and secret.
	- a valid auth0 client id and secret.
	- a valid auth0 domain.
	- a cluster name

You must also be in the infrastructure repository, and have decrypted the repository before running this command.


```
cloud-platform cluster delete [flags]
```

### Options

```
      --auth0-client-id string       [required] auth0 client id to use
      --auth0-client-secret string   [required] auth0 client secret to use
      --auth0-domain string          [required] auth0 domain to use
      --destroy-cluster              [optional] if true, will destroy the eks cluster (default true)
      --destroy-components           [optional] if true, will destroy the cluster components (default true)
      --destroy-core                 [optional] if true, will destroy the cluster core layer (default true)
      --destroy-vpc                  [optional] if true, will destroy the vpc (default true)
      --dry-run                      [optional] if false, the cluster will be destroyed otherwise no changes will be made to the cluster (default true)
  -h, --help                         help for delete
      --kubecfg string               [optional] path to kubeconfig file (default "/home/runner/.kube/config")
      --name string                  [required] name of the cluster
      --terraform-version string     [optional] the terraform version to use. (default "1.2.5")
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform cluster](cloud-platform_cluster.md)	 - Cloud Platform cluster actions

