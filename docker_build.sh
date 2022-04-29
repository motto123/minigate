#!/usr/bin/env bash

SEVER_DIR="server"

function build() {
  if [ "$1" == "log" ]; then
    return
  fi
  docker build --build-arg dirname="$1" -t "$1"_service -f ./server/"$1"/Dockerfile .
  echo ">>>>> docker building $1 <<<<<<"
  return
}

args="$*"
if [ $# -eq 0 ]; then
  if [ ! -d "$SEVER_DIR" ]; then
    echo "======= $SEVER_DIR isn't exist"
    exit 0
  fi

  fileNames=$(ls -l $SEVER_DIR | awk '{print $9}')
  args=$fileNames
fi

for str in $args; do
  build "$str"
done
