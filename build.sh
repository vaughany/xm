#!/bin/bash

echo -e "\e[1mBuilding Linux...\e[0m"
# env GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -a -o bin/xm .
# Static compilation, 'cos the newer (22.04) version of libc is not available on 18.04.
env GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w -linkmode external -extldflags "-static"' -a -o bin/xm .

echo -e "\e[1mDone.\n\e[0m"
ls -hl bin/

echo
file bin/*