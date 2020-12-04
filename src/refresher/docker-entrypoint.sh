#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

REFRESH_RATE=25600

while true; do
    /api/refresher_bin
    
    echo "Slepping ${REFRESH_RATE}"
    sleep $REFRESH_RATE
done
