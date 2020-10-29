#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

REFRESH_RATE=25600

while true; do
    env python3 /refresher/main.py
    
    echo "Slepping ${REFRESH_RATE}"
    sleep $REFRESH_RATE
done
