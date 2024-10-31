## cloud-platform terraform apply

Terraform apply.

```
cloud-platform terraform apply [flags]
```

### Options

```
      --aws-access-key-id string       [required] Access key id of service account to be used by terraform
      --aws-region string              [required] aws region to use
      --aws-secret-access-key string   [required] Secret access key of service account to be used by terraform
  -h, --help                           help for apply
      --is-pipeline                    [required] if the terraform is being executed from the pipeline
      --plan-filename string           [optional] the plan filename to be output from the terraform plan or used for the terraform apply eg. 'plan-$PR_NUM.out' [default] ''
      --redact                         Redact the terraform output before printing (default true)
      --terraform-version string       [optional] the terraform version to use. (default "1.2.5")
      --workdir string                 [optional] the terraform working directory to perform terraform operation [default] . (default ".")
  -w, --workspace string               [required] workspace where terraform is going to be executed
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform terraform](cloud-platform_terraform.md)	 - Terraform actions.

