#!/bin/bash

set -e

/Users/nerd/software/zokrates/bin/zokrates compile --debug -i circuits/root.zok

/Users/nerd/software/zokrates/bin/zokrates compute-witness -a $1 $2 $3 $4

/Users/nerd/software/zokrates/bin/zokrates generate-proof
