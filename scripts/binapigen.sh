#!/bin/bash

API_DIR=${1}
if [ -z ${API_DIR} ]; then
  echo "ERROR: API directory not specified"
  exit 1
fi

binapi-generator -gen="" --output-dir=./stats/local/binapi --input-file=${API_DIR}/core/interface.api.json
binapi-generator -gen="" --output-dir=./stats/local/binapi --input-file=${API_DIR}/core/ip.api.json
binapi-generator --output-dir=./stats/local/binapi --input-file=${API_DIR}/core/vpe.api.json
binapi-generator -gen="" --output-dir=./stats/local/binapi --input-file=${API_DIR}/plugins/dhcp.api.json