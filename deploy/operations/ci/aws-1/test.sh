#!/usr/bin/env bash

set -eo pipefail

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}" || exit 1

clean () {
  echo "Cleaning infrastructure"
  terraform destroy -auto-approve
}

terraform init
clean
terraform plan
#terraform apply -auto-approve
clean

