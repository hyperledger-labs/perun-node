#!/bin/sh

set -e

if [ ! -f "build/setEnv.sh" ]; then
    echo "This script can only be ran from the root repository."
    exit 2
fi

# execute go run build/ci.go with commands
 exec "$@"
