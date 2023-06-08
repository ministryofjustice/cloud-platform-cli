## cloud-platform decode-secret

Decode a kubernetes secret

```
cloud-platform decode-secret [flags]
```

### Examples

```
$ cloud-platform decode-secret -n mynamespace -s mysecret
$ cloud-platform decode-secret -s mysecret  [if you are setting namespace via kubectl context]
	
```

### Options

```
  -e, --export-aws-credentials   Export AWS credentials as shell variables
  -h, --help                     help for decode-secret
  -n, --namespace string         Namespace name
  -r, --raw                      Output the raw secret, rather than prettyprinting
  -s, --secret string            Secret name
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform](cloud-platform.md)	 - Multi-purpose CLI from the Cloud Platform team

