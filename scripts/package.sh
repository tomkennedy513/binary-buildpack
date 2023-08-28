#!/usr/bin/env bash

set -eu
set -o pipefail

cd $(dirname "${BASH_SOURCE[0]}")/..

function package() {
  image=$1
  version=$2
  publish=$3

  buildpack_toml_path=$(dirname $0)/../buildpack.toml
  buildpackage_path=$(mktemp -d)
  mkdir -p $buildpackage_path
  trap "rm -rf $buildpackage_path" EXIT

  jam pack \
    --buildpack $buildpack_toml_path \
    --offline \
    --version $version \
    --output $buildpackage_path/buildpack.tgz

  cat > $buildpackage_path/package.toml <<EOF
  [buildpack]
  uri = "buildpack.tgz"
EOF

  publish_arg=""
  if "${publish}"; then
    publish_arg="--publish"
  fi
  pack buildpack package $image --config=$buildpackage_path/package.toml --format=image $publish_arg
}


function usage() {
  cat <<-USAGE
package.sh [OPTIONS]

Packages the binary buildpack.

Prerequisites:
- pack and jam installed
- docker login to your registry (if publishing)

OPTIONS
  --help                          -h  prints the command usage
  --image <image>                 buildpack image to build (e.g. gcr.io/myproject/my-repo or my-dockerhub-username/repo)
  --version <version>             version of buildpack
  --publish                       (boolean) publish buildpack to remote registry
USAGE
}

function main() {
  local image version publish
  publish="false"

  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --help|-h)
        shift 1
        usage
        exit 0
        ;;

      --image)
        image=("${2}")
        shift 2
        ;;

      --version)
        version=("${2}")
        shift 2
        ;;

      --publish)
        publish="true"
        shift 1
        ;;

      "")
        # skip if the argument is empty
        shift 1
        ;;

      *)
        echo -e "unknown argument \"${1}\"" >&2
        exit 1
    esac
  done

 if [ -z "${image:-}" ]; then
  echo "--image is required"
  usage
  exit 1
 fi

 if [ -z "${version:-}" ]; then
   echo "--version is required"
   usage
   exit 1
 fi

 package $image $version $publish
}

main "${@:-}"