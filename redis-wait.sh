#!/bin/sh

# Script: redis-wait.sh
# Author: Jess Bodzo
# Usage redis-wait.sh $1 $2..n
# Args: 
#   $1 - redis endpoint you wish to connect to
#   $2..n - all the args to pass to the app container
# Purpose:
#   Docker Compose will not wait for redis to finish
#   configuration before launching our application container
#   that depends on it. This makes sure we wait until redis
#   is ready before trying to launch our container.

set -e
endpoint=$1
shift
cmd=$@
MAX_ATTEMPTS=5
sleep 4

checks=0
>&2 echo "Redis ready check..."
while [ "$checks" -lt "$MAX_ATTEMPTS" ]; do
    redisResp="$(redis-cli -h $endpoint ping)"
    if [ "$redisResp" == "PONG" ]; then
        >&2 echo "Redis ready"
        sleep 2
        break
    else
        >&2 echo "Redis not ready yet"
        checks=$((checks+1))
        sleep 1
    fi
done

exec $cmd