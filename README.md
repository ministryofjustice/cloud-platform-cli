# Cloud Platform Tool CLI

`cloud-platform` is a command-line tool used by the cloud-platform team to achieve common repetitive tasks against the platform. The resulting binary checks things like:

   - Create templates for the environments repo
   - Divergences in terraform states
   - Terraform apply
   - Others

## Install

`cloud-platform` can be installed and upgraded by running:

```shell
GO111MODULE=on go get github.com/ministryofjustice/cloud-platform-cli/cmd/cloud-platform
```

## Usage

`cloud-platform` has different subcommands. Execute: `cloud-platform --help` in order to check them out. Remember some of the subcommands requires `AWS_*` keys.

### `cloud-platform terraform`

`cloud-platform terraform` simply is a terraform wrapper which allows us to execute terraform commands. It also includes `check-divergence` command which fails in case there are pending terraform changes.

Usage example:

```shell
$ cloud-platform terraform check-divergence --workspace manager --var-file prod.tfvars
INFO[0000] Executing terraform plan, if there is a drift this program execution will fail  subcommand=check-divergence
INFO[0002] Initializing modules...

Initializing the backend...

Initializing provider plugins...

...
...
...
Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
INFO[0003]
INFO[0015] Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

tls_private_key.worker_key: Refreshing state... [id=1fdf4d57aa5f06668ce94fd60c2a5de657d07de4]
tls_private_key.session_signing_key: Refreshing state... [id=66e5f736b76f51d26d37c1449b21e89066f369d7]
tls_private_key.host_key: Refreshing state... [id=4bf29261ec098ddf34c7d3a45d3b460bd6585ed5]
...
...
...

kubernetes_namespace.concourse: Refreshing state... [id=concourse]
kubernetes_namespace.concourse_main: Refreshing state... [id=concourse-main]
kubernetes_secret.concourse_aws_credentials: Refreshing state... [id=concourse-main/aws-manager]
kubernetes_secret.concourse_basic_auth_credentials: Refreshing state... [id=concourse-main/concourse-basic-auth]

------------------------------------------------------------------------

No changes. Infrastructure is up-to-date.

This means that Terraform did not detect any differences between your
configuration and real physical resources that exist. As a result, no
actions need to be performed.

```

## Develop

You will need golang installed (version 1.14 or greater).

### Build locally

```
go mod download
go build -o cloud-platform ./cmd/cloud-platform/main.go
```

This will create a `cloud-platform` binary.

### Updating / Publishing

This project includes a [github action](.github/workflows/docker-hub.yml) which
will automatically build a new docker image and push it to [docker hub], tagged
with the release number, whenever you create a new release via the [github ui].

[docker hub]: https://hub.docker.com/repository/docker/ministryofjustice/cloud-platform-cli
[github ui]: https://github.com/ministryofjustice/cloud-platform-cli/releases
