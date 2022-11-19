#!/bin/bash

set -e

LOCAL_INST_DIR="$HOME/.local/bin"
BIN_NAME="zokrates"

mkdir -p $LOCAL_INST_DIR

#curl -Ls https://scrypt.io/scripts/setup-zokrates.sh | bash

if [ -f "$LOCAL_INST_DIR/$BIN_NAME" ]; then
    echo "zokrates found"

    cd circuits

    $LOCAL_INST_DIR/zokrates compile --debug -i root.zok

    $LOCAL_INST_DIR/zokrates setup

    $LOCAL_INST_DIR/zokrates export-verifier-scrypt -o ../contract/verifier.scrypt

    exit 0
else
    echo "zokrates not found"
    exit 0
fi
