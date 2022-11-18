#!/bin/bash

set -e

cd circuits

/Users/nerd/software/zokrates/bin/zokrates compile --debug -i root.zok

/Users/nerd/software/zokrates/bin/zokrates setup

/Users/nerd/software/zokrates/bin/zokrates export-verifier-scrypt -o ../verifier.scrypt
