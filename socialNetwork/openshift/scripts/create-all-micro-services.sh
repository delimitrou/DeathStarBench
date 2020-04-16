
cd $(dirname $0)/..

NS="social-network"

oc create namespace ${NS}

for service in *service*.yaml ; do
  oc apply -f $service -n ${NS}
done

cd - >/dev/null
