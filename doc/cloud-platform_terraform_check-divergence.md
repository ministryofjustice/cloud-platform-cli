## cloud-platform terraform check-divergence

Terraform check-divergence check if there are drifts in the state.

```
cloud-platform terraform check-divergence [flags]
```

### Options

```
      --aws-access-key-id string       Access key id of service account to be used by terraform
      --aws-region string              [required] aws region to use
      --aws-secret-access-key string   Secret access key of service account to be used by terraform
  -h, --help                           help for check-divergence
      --redact                         Redact the terraform output before printing (default true)
      --terraform-version string       [optional] the terraform version to use. [default] 0.14.8 (default "0.14.8")
      --workdir string                 [optional] the terraform working directory to perform terraform operation [defaukt] . (default ".")
  -w, --workspace string               Default workspace where terraform is going to be executed (default "default")
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform terraform](cloud-platform_terraform.md)	 - Terraform actions.

