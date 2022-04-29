#! /bin/bash
#build and restart server
SCRIPT_DIR=$(
  cd $(dirname ${BASH_SOURCE[0]})
  pwd
)

cd "$SCRIPT_DIR"

#if [ $# -eq 0 ]; then
#  echo usage: ./br.sh auth
#  echo usage: ./br.sh auth gate ...
#  exit 1
#fi
#

SEVER_DIR="server"

args="$@"
if [ $# -eq 0 ]; then
  if [ ! -d "$SEVER_DIR" ]; then
    echo "======= $SEVER_DIR isn't exist"
    exit 0
  fi

  fileNames=$(ls -l server | awk '{print $9}')
  args=$fileNames
fi

for param in $args; do
  # 将脚本的错误输出到文件中，判断文件是否空,如果为空继续执行,否则return
  ./build.sh "$param" 2>/tmp/builderr.txt
  if [ -z "$(cat /tmp/builderr.txt)" ]; then
    ./restart.sh "$param"
    echo "=============================="
  else
    # 把错误交给标准错误
    cat /tmp/builderr.txt | xargs echo >&2
  fi

done

rm /tmp/builderr.txt 2>/dev/null
