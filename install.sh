#!/bin/bash

{
  set -u

  if [ -d "/usr/local/Cellar/cloud-platform-cli" ]; then
    echo "cloud-platform-cli already installed via Homebrew"
    cloud-platform version
    exit
  fi

  cached_uname=($(uname -sm))
  kernel_name="${cached_uname[0]}"
  machine_arch="${cached_uname[1]}"

  case $kernel_name in
    Darwin)
      os="darwin" ;;
    Linux|GNU*)
      os="linux" ;;
    *)
  esac

  case $machine_arch in
    x86_64)
      arch="amd64" ;;
    arm64)
      arch="arm64" ;;
    *)
  esac

  if [ -z ${os+x} ]; then
    echo "Unsupported OS: this installation script only supports darwin, linux"
    exit
  fi

  if [ -z ${arch+x} ]; then
    echo "Unsupported architecture: this installation script only supports amd64, arm64"
    exit
  fi

  # Get os/arch download URL
  latest_tag=$(curl -sL https://api.github.com/repos/ministryofjustice/cloud-platform-cli/releases/latest | jq -r '.tag_name')
  download_url="https://github.com/ministryofjustice/cloud-platform-cli/releases/download/${latest_tag}/cloud-platform-cli_${latest_tag}_${os}_${arch}.tar.gz"

  echo "Installing CLI from $download_url"

  # # Install into /usr/local/bin
  curl -sL "$url" | tar xz -C /usr/local/bin

  # Test the CLI
  version=$(cloud-platform version)
  echo "Cloud Platform $version installed!"
}
