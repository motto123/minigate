#! /bin/bash
#restart server

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

for d in ${args}; do
  d=${d%%/*}

  echo "killing $d"
  pkill -TERM "^$d$"

  sleep 1
  cd "$SEVER_DIR/$d/bin"
  ./$d &
  echo ">>>>>> restart $d <<<<<<"
  cd - >/dev/null

done

echo
