#!/usr/bin/env bash

SCRIPT_DIR=$(
  cd $(dirname ${BASH_SOURCE[0]})
  pwd
)
cd "$SCRIPT_DIR"

SEVER_DIR="server"

function build() {
  cd "$SEVER_DIR/$1"
  if [ ! -d "bin" ]; then
    mkdir bin
  fi
  rm "bin/$1" 2>/dev/null
  go mod tidy
  go build -o "./bin/$1" main.go
  echo ">>>>> building $1 <<<<<<"
  cd ../../
  return
}

args="$@"
if [ $# -eq 0 ]; then
  if [ ! -d "$SEVER_DIR" ]; then
    echo "======= $SEVER_DIR isn't exist"
    exit 0
  fi

  fileNames=$(ls -l server | awk '{print $9}')
  args=$fileNames
fi

for str in $args; do
  build "$str"
done
