# Cloud Platform Tool CLI

[![Releases](https://img.shields.io/github/release/ministryofjustice/cloud-platform-cli/all.svg?style=flat-square)](https://github.com/ministryofjustice/cloud-platform-cli/releases)
[![codecov](https://codecov.io/gh/ministryofjustice/cloud-platform-cli/branch/main/graph/badge.svg?token=BUF45279MY)](https://codecov.io/gh/ministryofjustice/cloud-platform-cli)

`cloud-platform` is a command-line tool used by the cloud-platform team and tenants to perform actions on the platform, for example:

- Create environment configuration using a template
- Divergences in terraform states
- Terraform apply
- Others

User documentation is here: https://user-guide.cloud-platform.service.justice.gov.uk/documentation/getting-started/cloud-platform-cli.html

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

## Develop

You will need Go installed.

### Build locally

Run `make` to create a `cloud-platform` binary.

[note] Something worth noting when building locally. You'll need to pass the `--skip-version-check` command to avoid a message about upgrading.

### Testing

Run `make test` to run the unit tests.

### Updating / Publishing

This project includes a [github action](.github/workflows/build-release.yml) which
will automatically do the following steps:

- build a new release and make it available in the [github ui]
- build a new docker image and push it to [docker hub], tagged with the version number

In order to trigger this action, push a new tag version like this:

```
git tag [my new version]
git push --tags
```

The value of this tag **must** be built into the binary `Version` in the file `pkg/commands/version.go`. This will happen automatically on release.

#### `PreRun` hook

**Every** new command should have a PreRun hook as follows, to ensure the version of the cli tool is consistent:

```
PreRun: upgradeIfNotLatest,
```

See the existing commands for examples.

[docker hub]: https://hub.docker.com/repository/docker/ministryofjustice/cloud-platform-cli
[github ui]: https://github.com/ministryofjustice/cloud-platform-cli/releases
