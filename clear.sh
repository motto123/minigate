#!/usr/bin/env bash

SCRIPT_DIR=$(cd $(dirname ${BASH_SOURCE[0]}); pwd)

cd "$SCRIPT_DIR"

echo ">>>>>> start clear server bin <<<<<<"
rm -rf server/*/bin/*
echo ">>>>>> start clear all xxx.log in project <<<<<<"
rm -rf log/*

echo ">>>>>> done <<<<<<"
