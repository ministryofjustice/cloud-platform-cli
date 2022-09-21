## cloud-platform duplicate ingress

Duplicate ingress for the given ingress resource name and namespace

### Synopsis

Gets the ingress resource for the given name and namespace from the cluster,
copies it, change the ingress name and external-dns annotations for the weighted policy and
apply the duplicated ingress to the same namespace.

This command will access the cluster to get the ingress resource and to create the duplicate ingress.
To access the cluster, it assumes that the user has either set the env variable KUBECONFIG to the filepath of kubeconfig or stored the file in the default location ~/.kube/config
	

```
cloud-platform duplicate ingress <ingress name> [flags]
```

### Examples

```
$ cloud-platform duplicate ingress myingressname -n mynamespace


```

### Options

```
  -h, --help               help for ingress
  -n, --namespace string   Namespace which you want to perform the duplicate resource
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform duplicate](cloud-platform_duplicate.md)	 - Cloud Platform duplicate resource

