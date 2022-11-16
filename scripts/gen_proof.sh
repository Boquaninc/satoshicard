#!/bin/bash

set -e

cd circuits

/Users/linxing/.local/bin/zokrates compute-witness -a $1 $2 $3 $4

/Users/linxing/.local/bin/zokrates generate-proof
