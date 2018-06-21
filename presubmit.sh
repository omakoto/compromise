#!/bin/bash

# Extract commands from .travis.yml and execute them.

set -e

cd "${0%/*}"

run() {
  echo "Running: $*"
  "$@"
}

. <(sed -ne 's/^ *- *\(.*\)#presubmit/run \1/p' .travis.yml)
