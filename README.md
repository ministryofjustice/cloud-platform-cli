# Cloud Platform Tool CLI

[![Releases](https://img.shields.io/github/release/ministryofjustice/cloud-platform-cli/all.svg?style=flat-square)](https://github.com/ministryofjustice/cloud-platform-cli/releases)

`cloud-platform` is a command-line tool used by the cloud-platform team and tenants to perform actions on the platform, for example:

   - Create templates for the environments repo
   - Divergences in terraform states
   - Terraform apply
   - Others

## Install

### Homebrew

```
brew install ministryofjustice/cloud-platform-tap/cloud-platform-cli
```

### Manual

These installation instructions are for a Mac. If you have a different kind of
computer, please amend the steps appropriately.

Please substitute the latest release number. You can see the latest release
number in the badge near the top of this page, and all available releases on
[this page][github ui].

```
RELEASE=<insert latest release>
wget https://github.com/ministryofjustice/cloud-platform-cli/releases/download/${RELEASE}/cloud-platform-cli_${RELEASE}_darwin_amd64.tar.gz
tar xzvf cloud-platform-cli_${RELEASE}_darwin_amd64.tar.gz
mv cloud-platform /usr/local/bin/
```

NB: You may need to manually open the file to override OSX restrictions against
executing binaries downloaded from the internet. To do this, locate the file in
the Finder, right-click it and choose "Open". After doing this once, you should
be able to run the command as normal.

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

You will need golang installed (version 1.13 or greater).

### Build locally

Run `make` to create a `cloud-platform` binary.

### Testing

Run `make test` to run the unit tests.

### Updating / Publishing

This project includes a [github action](.github/workflows/build-release.yml) which
will automatically do the following steps:

* build a new release and make it available in the [github ui]
* build a new docker image and push it to [docker hub], tagged with the version number

In order to trigger this action, push a new tag version like this:

```
git tag [my new version]
git push --tags
```

The value of this tag **must** be the same as the string value of `Version` in the file `pkg/commands/version.go`

#### Self-upgrade `PreRun` hook

**Every** new command should have a PreRun hook as follows, to ensure the "self-upgrading" behaviour of the cli tool is consistent:

```
PreRun: upgradeIfNotLatest,
```

See the existing commands for examples.

[docker hub]: https://hub.docker.com/repository/docker/ministryofjustice/cloud-platform-cli
[github ui]: https://github.com/ministryofjustice/cloud-platform-cli/releases
