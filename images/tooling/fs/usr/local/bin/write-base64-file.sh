#!/usr/bin/env sh

echo "$1" | base64 -d - > "$2"
