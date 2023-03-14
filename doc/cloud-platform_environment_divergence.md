## cloud-platform environment divergence

Check for divergence between the environments repository and the cluster

```
cloud-platform environment divergence [flags]
```

### Examples

```
> cloud-platform environment divergence --cluster myTestCluster --githubToken myGithubToken123

```

### Options

```
  -c, --cluster-name string   [optional] Cluster name (default "live")
  -g, --github-token string   [required] Github token
  -h, --help                  help for divergence
  -k, --kubeconfig string     [optional] Kubeconfig file path
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform environment](cloud-platform_environment.md)	 - Cloud Platform Environment actions

