#!/bin/sh

set -e

apparea serve \
    --bind-ssh ${SSH_ADDRESS:="0.0.0.0:21"} \
    --bind-http ${HTTP_ADDRESS:="0.0.0.0:80"} \
    --hostname ${DOMAIN:=apparea.localhost}
