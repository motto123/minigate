#! /bin/bash
#kill
SCRIPT_DIR=$(cd $(dirname ${BASH_SOURCE[0]}); pwd)
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
  echo "======== killing all server ...... ========="
fi


for d in $args; do
  pid=$(ps -ef | grep "./${d}$" | awk '{print $2}')
  kill -9 $pid 2>> /dev/null
  if [[ "$?" -eq "0" ]];then
    echo ">>>>>> killed $d <<<<<<<"
    else
    echo ">>>>>> kill failed $d <<<<<<<"
  fi
done
