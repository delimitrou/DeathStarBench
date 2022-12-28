#!/bin/bash


cd $(dirname $0)/../../

istioIngress=$(kubectl get svc -n istio-system | grep "istio-ingressgateway" | awk '{print $3}')
namespace=default.
domainName=.sslip.io:80


if [ `grep -c "KnativeDomainName" "config.json"` -ne '0' ];then
    echo "Append config: KnativeDomainName"
    sed -i '/"KnativeDomainName"/c\  \"KnativeDomainName\": \"'$namespace$istioIngress$domainName'\"'  config.json
    exit 0
fi

sed -i '/"UserMongoAddress"/ s/$/&,/' config.json
sed -i '$i\  "KnativeDomainName\": \"'$namespace$istioIngress$domainName'\"'  config.json



