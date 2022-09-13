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

### via Homebrew

```
brew install ministryofjustice/cloud-platform-tap/cloud-platform-cli
```

### Manually

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

The `/doc` directory should contain usage instructions, otherwise, please see the output of `cloud-platform --help` or the [user-guide](https://user-guide.cloud-platform.service.justice.gov.uk/documentation/getting-started/cloud-platform-cli.html) entry for more information.

### Autogenerate documentation

The cli uses the [cobra-docs](https://github.com/spf13/cobra/blob/main/doc/md_docs.md) generator to create automated Markdown pages from Cobra.

When a pull-request is opened, a GitHub Action will trigger and autogenerate the documentation. The action will commit these changes back to the remote branch.

## Develop

You will need Go installed.

### Build locally

Run `make` to create a `cloud-platform` binary.

[note] Something worth noting when building locally. You'll need to pass the `--skip-version-check` command to avoid a message about upgrading.

### Testing

There are two types of tests in this repository:

#### Integration

These tests build the root binary and test the output of a command. For example, `cloud-platform version` will output `testBuild` using a package called [go-testcmd](https://github.com/google/go-cmdtest). Integration tests are tagged with `integration` so won't run using the normal `go test -v ./...` command. You'll have to pass the `integration` keyword as a build tag, i.e. `go test -v ./... --tags integration`

If you'd like to create a new integration test, add the following to the top of your test file: `//go:build integration`.

If the output of a command changes and the tests start failing, simply add the `-update` flag to your test command and they'll automatically update on your behalf. For example: `go test . --tags integration -update`

#### Unit

These tests live next to the code, they have no build tag and will run regardless of the flag you on build.

Run `make test` to run the unit tests.

There are Dockerfile structure tests that run automatically in a pipeline. If you want to run these locally, install the [container-structure-test](https://github.com/GoogleContainerTools/container-structure-test#installation) binary and run:

```bash
container-structure-test test --image my-image-name \
--config docker-test.yaml
```

### Releasing a new version

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
