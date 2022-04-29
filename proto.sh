#!/usr/bin/env bash

SCRIPT_DIR=$(
  cd $(dirname ${BASH_SOURCE[0]})
  pwd
)

cd "$SCRIPT_DIR"

cd "proto" 2>/dev/null || exit

errOutFile="/tmp/protoErr.txt"
package=$(cat go.mod | grep module | awk '{print $2}')

rm $errOutFile 2>/dev/null

function autoRepairDependence() {
  for filename in $(ls -l | grep ^d | awk '{print $9}'); do
    sed -i 's/"\/'"${filename}"'"/"'"${package}"'\/'"${filename}"'"/g' "$1"/*.pb.go
    if [ $? != 0 ]; then
      echo "autoRepairDependence failed" >>$errOutFile
      return
    fi
  done
}

function generalProto() {
  #  rm "$1"/*.pb.go 2>/dev/null
  protoc --proto_path=. --go_out=plugins=grpc:. ./"$1"/*.proto 2>>$errOutFile
  #  protoc --proto_path=. --go_out=. ./$1/*.proto
  # 修复引用common/xxx.proto后import路径报错
  autoRepairDependence $1
  if [[ -s $errOutFile ]]; then
    return
  fi

  protoc-go-inject-tag -input="./$1/*.pb.go" 2>>$errOutFile

  echo ">>>>> general $1 proto file <<<<<<<<"
}

args=$@
if [ $# -eq 0 ]; then
  args=$(ls -l | grep ^d | awk '{print $9}')
fi

for filename in $args; do
  info=$(generalProto "$filename")

#  if [ ! -z "$(cat $errOutFile)" ]; then
  if [ -s $errOutFile ]; then
    # 把错误交给标准错误
    echo ">>>>> general $filename proto file failed <<<<<<<<"
    cat $errOutFile | xargs echo >&2
    exit 0
  fi
  echo "$info"

done
