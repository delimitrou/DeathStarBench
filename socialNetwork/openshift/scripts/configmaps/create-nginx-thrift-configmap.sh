#!/bin/bash

NS="social-network"

cd $(dirname $0)/../..

# directories

oc create cm nginx-thrift-genlua --from-file=nginx-thrift-config/gen-lua -n ${NS}
oc create cm nginx-thrift-pages  --from-file=nginx-thrift-config/pages -n ${NS}
oc create cm nginx-thrift-pages-style --from-file=nginx-thrift-config/pages/style -n ${NS}
oc create cm nginx-thrift-pages-javascript  --from-file=nginx-thrift-config/pages/javascript -n ${NS}
oc create cm nginx-thrift-luascripts --from-file=nginx-thrift-config/lua-scripts -n ${NS}
oc create cm nginx-thrift-luascripts-api-home-timeline --from-file=nginx-thrift-config/lua-scripts/api/home-timeline -n ${NS}
oc create cm nginx-thrift-luascripts-api-post --from-file=nginx-thrift-config/lua-scripts/api/post -n ${NS}
oc create cm nginx-thrift-luascripts-api-user --from-file=nginx-thrift-config/lua-scripts/api/user -n ${NS}
oc create cm nginx-thrift-luascripts-api-user-timeline --from-file=nginx-thrift-config/lua-scripts/api/user-timeline -n ${NS}
oc create cm nginx-thrift-luascripts-wrk2-api-home-timeline --from-file=nginx-thrift-config/lua-scripts/wrk2-api/home-timeline -n ${NS}
oc create cm nginx-thrift-luascripts-wrk2-api-post --from-file=nginx-thrift-config/lua-scripts/wrk2-api/post -n ${NS}
oc create cm nginx-thrift-luascripts-wrk2-api-user --from-file=nginx-thrift-config/lua-scripts/wrk2-api/user -n ${NS}
oc create cm nginx-thrift-luascripts-wrk2-api-user-timeline --from-file=nginx-thrift-config/lua-scripts/wrk2-api/user-timeline -n ${NS}

# file
oc create cm nginx-thrift        --from-file=nginx-thrift-config/nginx.conf -n ${NS}


# nginx-thrift has a dependency on the ConfigMap nginx-thrift-jaeger, which
# is created by the script `create-jaeger-configmap.sh`
