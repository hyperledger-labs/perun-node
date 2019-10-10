#!/bin/sh

set -e

if [ ! -f "build/setEnv.sh" ]; then
    echo "This script can only be ran from the root repository."
    exit 2
fi

    echo "Creating a fake go workspace, so that user can build from source placed in non-go directory" 
    tempGoPath="$PWD/build/workspace_"

    #Any change in project path should be reflected in DST_GO_PATH and tempDir
    DST_GO_PATH="/src/github.com/direct-state-transfer/dst-go/"
    export DST_GO_PATH
    root="$PWD"
    tempDir="$tempGoPath/src/github.com/direct-state-transfer"

    if [ ! -L "$tempDir/dst-go" ]; then
     mkdir -p "$tempDir"
     cd "$tempDir"
     ln -s ../../../../../. dst-go
     cd "$root"
    fi
   
    # Set up the environment to use the workspace.
    GOPATH="$tempGoPath"
    export GOPATH

    # Change Dir
    cd "$tempDir/dst-go"
    PWD="$tempDir/dst-go"   
    echo "Temp GOPATH set to fake go workspace : $PWD \n"

# execute go run build/ci.go with commands
 exec "$@"
