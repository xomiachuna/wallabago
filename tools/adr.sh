#!/bin/bash

docker run --rm -it \
    --name adr-tools \
    -v `pwd`/docs:/doc \
    rdhaliwal/adr-tools \
    adr "$@"
