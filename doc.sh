#!/usr/bin/env bash

SCRIPT_DIR=$(cd $(dirname ${BASH_SOURCE[0]}); pwd)

cd "$SCRIPT_DIR"

src="doc/doc.md"
dist="doc/doc.html"

rm ${dist} 2> /dev/null
pandoc --toc-depth=4 --toc -s -c ./resource/doc_css/vue.css --self-contained -f markdown -t html ${src} -o ${dist}

echo "visit local file://$(pwd)/${dist} by the website"


