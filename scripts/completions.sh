#!/bin/sh
# scripts/completions.sh
# Creates completions for goreleaser.
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run ./main.go completion "$sh" >"completions/cloud-platform.$sh"
done
