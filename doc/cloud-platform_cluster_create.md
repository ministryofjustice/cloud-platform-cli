## cloud-platform cluster create

Create a new Cloud Platform cluster

```
cloud-platform cluster create [flags]
```

### Examples

```
$ cloud-platform cluster create --name my-cluster

```

### Options

```
      --auth0-client-id string       [required] auth0 client id to use
      --auth0-client-secret string   [required] auth0 client secret to use
      --auth0-domain string          [required] auth0 domain to use
      --cluster-suffix string        [optional] suffix to append to the cluster name (default "cloud-platform.service.justice.gov.uk")
      --debug                        [optional] enable debug logging
      --fast                         [optional] enable fast mode - this creates a cluster as quickly as possible. [default] false
  -h, --help                         help for create
      --name string                  [optional] name of the cluster (default "jb-2309-1719")
      --nodes int                    [optional] number of nodes to create. [default] 3 (default 3)
      --terraformVersion string      [optional] the terraform version to use. [default] 0.14.8 (default "0.14.8")
      --vpc string                   [optional] name of the vpc to use (default "jb-2309-1719")
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform cluster](cloud-platform_cluster.md)	 - Cloud Platform cluster actions

