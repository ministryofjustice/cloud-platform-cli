## cloud-platform environment rds-drift-checker

Detect and correct RDS engine version drift from a CSV file in S3 or locally

```
cloud-platform environment rds-drift-checker <file-location> [flags]
```

### Examples

```
Run with a file from S3:
  cloud-platform environment rds-drift-checker s3://your-bucket/path/to/merged-rds-errored-namespaces.csv

Run with a local CSV file:
  cloud-platform environment rds-drift-checker file://path/to/merged-rds-errored-namespaces.csv

```

### Options

```
  -h, --help   help for rds-drift-checker
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform environment](cloud-platform_environment.md)	 - Cloud Platform Environment actions

