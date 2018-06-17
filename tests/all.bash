#!/bin/bash

#set -e

cd "${0%/*}/.."

for n in tests/test*.bash ; do
    echo "# $n"
    "$n"
done