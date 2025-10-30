#!/bin/sh

if test -z "$1" || test "$1" = "suns"; then
    endpoint="https://suns.bz/api/v1/attest"
else
    endpoint="https://pz39hzelmj.execute-api.us-east-2.amazonaws.com/api/v1/attest"
fi

set -x

curl -X POST "$endpoint" \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "https://me.micahrl.com",
    "type": "mirrornames",
    "domains": ["me.micahrl.com", "com.micahrl.me"]
  }'
