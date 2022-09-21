## cloud-platform cluster recycle-node

recycle a node

```
cloud-platform cluster recycle-node [flags]
```

### Examples

```
$ cloud-platform cluster recycle-node

```

### Options

```
      --aws-region string   aws region to use (default "eu-west-2")
      --debug               enable debug logging
  -f, --force               force the pods to drain (default true)
  -h, --help                help for recycle-node
  -i, --ignore-label        whether to ignore the labels on the resource
      --kubecfg string      path to kubeconfig file (default "/home/runner/.kube/config")
  -n, --name string         name of the resource to recycle
      --oldest              whether to recycle the oldest node
  -t, --timeout int         amount of time to wait for the drain command to complete (default 360)
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform cluster](cloud-platform_cluster.md)	 - Cloud Platform cluster actions

