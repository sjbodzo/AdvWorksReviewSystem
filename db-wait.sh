#!/bin/sh

# Script: db-wait.sh
# Author: Jess Bodzo
# Usage db-wait.sh $1 $2..n
# Args: 
#   $1 - db host you wish to connect to
#   $2..n - all the args to pass to the app container
# Purpose:
#   Docker Compose will not wait for the database to finish
#   configuration before launching our application container
#   that depends on it. This makes sure we wait until the db
#   is ready before trying to launch our container.
# Dependencies:
#   Modifying the sql script for the database could break
#   the script, because it could change what it means for
#   the database to be ready for us.

set -e
db=$1
shift
cmd=$@
MAX_ATTEMPTS=10
sleep 3

checks=0
>&2 echo "DB ready check..."
while [ "$checks" -lt "$MAX_ATTEMPTS" ]; do
    schemaCount=`echo "SELECT COUNT(*) from information_schema.tables" | psql -qtAX "dbname=AdventureWorks host=$db user=postgres password=postgres"`
    if [ "$schemaCount" == "343" ]; then
        reviewCount=`echo "SET search_path=production; SELECT COUNT(*) FROM Production.ProductReview" | psql -qtAX "dbname=AdventureWorks host=$db user=postgres password=postgres"`
        if [ "$reviewCount" == "5" ]; then
            >&2 echo "DB ready"
            break
        else
            >&2 echo "DB not ready yet"
            checks=$((checks+1))
            sleep 10
        fi
    else
        >&2 echo "DB not ready yet"
        checks=$((checks+1))
        sleep 10
    fi
done

exec $cmd