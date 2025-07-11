#!/bin/bash

docker run --rm \
    -v `pwd`/docs:/doc \
    rdhaliwal/adr-tools \
    adr "$@"
