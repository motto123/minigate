#!/usr/bin/env bash

SCRIPT_DIR=$(
  cd $(dirname ${BASH_SOURCE[0]})
  pwd
)
SEVER_DIR="server"

cd "$SCRIPT_DIR"

args="$@"
if [ $# -eq 0 ]; then
  if [ ! -d "$SEVER_DIR" ]; then
    echo "======= $SEVER_DIR isn't exist"
    exit 0
  fi

  fileNames=$(ls -l server | awk '{print $9}')
  args=$fileNames
fi

function build() {
  cd "$SEVER_DIR/$1"
  if [ ! -d "bin" ]; then
    mkdir bin
  fi
  rm "bin/$1" 2>/dev/null
  go mod tidy
  go build -gcflags=all="-N -l" -o "./bin/$1" main.go
  echo ">>>>> building $1 <<<<<<"
  cd ../../
}

function debug() {
  if [ $# -ne 0 ]; then
    n=6600
    for d in $@; do
      local d=${d%%/*}
      echo ">>> kill $d and dlv server <<<"
      ps -ef | grep -E "/${d}$" | awk '{print $2}' | xargs -r kill -9
      build $d
      tip=">>> debug $d, dlv server listening :$n  <<<"
      echo -e "\033[33m${tip}\033[0m"
      dlv --listen=":$n" --headless=true --api-version=2 exec "./server/$d/bin/$d" &
      n=$((n + 1))
    done
  else
    echo "please input params, example: ./build auth or ./build auth user ..."
  fi
}

debug $args
