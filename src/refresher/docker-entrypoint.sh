#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

REFRESH_RATE=259200

while true; do
    /api/refresher_bin -db-address database
    
    echo "Slepping ${REFRESH_RATE}"
    sleep $REFRESH_RATE
done
