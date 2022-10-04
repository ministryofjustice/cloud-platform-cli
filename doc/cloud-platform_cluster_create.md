## cloud-platform cluster create

create a new cloud-platform cluster

### Synopsis


Running this command will create a new eks cluster in the cloud-platform aws account.
It will create you a VPC, a cluster and the components the cloud-platform team defines as being required for a cluster to function.

You must have the following environment variables set, or passed via arguments:
	- a valid AWS profile or access key and secret.
	- a valid auth0 client id and secret.
	- a valid auth0 domain.

You must also be in the infrastructure repository, and have decrypted the repository before running this command.


```
cloud-platform cluster create [flags]
```

### Options

```
      --auth0-client-id string       [required] auth0 client id to use
      --auth0-client-secret string   [required] auth0 client secret to use
      --auth0-domain string          [required] auth0 domain to use
      --cluster-suffix string        [optional] suffix to append to the cluster name (default "cloud-platform.service.justice.gov.uk")
  -h, --help                         help for create
      --name string                  [optional] name of the cluster
      --terraform-version string     [optional] the terraform version to use. [default] 0.14.8 (default "0.14.8")
      --vpc string                   [optional] name of the vpc to use
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform cluster](cloud-platform_cluster.md)	 - Cloud Platform cluster actions

