#!/bin/bash

set -e

cd circuits

LOCAL_INST_DIR="$HOME/.local/bin"

BIN_NAME="zokrates"

zokrates compute-witness -a $1 $2 $3 $4

zokrates generate-proof
