#!/bin/bash

set -e

LOCAL_INST_DIR="$HOME/.local/bin"
BIN_NAME="zokrates"

mkdir -p $LOCAL_INST_DIR

#curl -Ls https://scrypt.io/scripts/setup-zokrates.sh | bash

#if [ -f "$LOCAL_INST_DIR/$BIN_NAME" ]; then
if ! command -v zokrates &> /dev/null
then
    echo "zokrates cmd not found."

    echo "exec 'curl -Ls https://scrypt.io/scripts/setup-zokrates.sh | bash' to install it."
    echo "or add 'alias zokrates=INSTALL_PATH to system env"
    exit 0
else
    echo "zokrates found"

    cd circuits

    zokrates compile --debug -i root.zok
    zokrates setup
    zokrates export-verifier-scrypt -o ../contract/verifier.scrypt

    #$LOCAL_INST_DIR/zokrates compile --debug -i root.zok
    #$LOCAL_INST_DIR/zokrates setup
    #$LOCAL_INST_DIR/zokrates export-verifier-scrypt -o ../contract/verifier.scrypt
fi
