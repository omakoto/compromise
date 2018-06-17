#!/bin/bash

# Simple smoke test -- make sure the output is readable by Bash and Zsh.

cd "${0%/*}/.."

export COMPROMISE_BASH_SKIP_BINDS=1

echo "Loading in bash..."
bash --norc --noprofile <(go run ./src/cmds/compromise-adb/adb.go) |& grep '^Installed completion'

echo "Loading in zsh..."
zsh --no-rcs <(go run ./src/cmds/compromise-adb/adb.go) |& grep '^Installed completion'