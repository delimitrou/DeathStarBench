#!/bin/bash

cd $(dirname $0)/../..

oc create cm media-frontend-nginx --from-file=media-frontend-config/nginx.conf  -n social-network
oc create cm media-frontend-lua   --from-file=media-frontend-config/lua-scripts -n social-network
