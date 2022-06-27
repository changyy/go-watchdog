#!/bin/bash

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]:-$0}"; )" &> /dev/null && pwd 2> /dev/null; )";
alias vim='vim -u $SCRIPT_DIR/vimrc'
echo "Update vim env for golang/php/python (using 'type vim' to show raw command) "
