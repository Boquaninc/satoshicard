#!/bin/bash

set -e

cd circuits

/Users/linxing/.local/bin/zokrates compile --debug -i root.zok

/Users/linxing/.local/bin/zokrates setup

/Users/linxing/.local/bin/zokrates export-verifier-scrypt -o ../verifier.scrypt
