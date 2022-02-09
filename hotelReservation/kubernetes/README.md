# Hotel Reservations on OpenShift 4.x

## Pre-requirements

- A running Kubernetes cluster is needed.
- Pre-requirements mentioned [here](https://github.com/delimitrou/DeathStarBench/blob/master/hotelReservation/README.md) should be met.

## Running the Hotel Reservation application

### Before you start

- Ensure that the necessary local images have been made.
  - `<path-of-repo>/hotelReservation/kubernetes/scripts/build-docker-images.sh`
  if you intend to change it, remember to change the username and image name in the build script and also all deployments as well.
### Deploy services

run `kubectl apply -Rf <path-of-repo>/hotelReservation/kubernetes/`
and wait for `kubectl get pods` to show all pods with status `Running`.


### Prepare HTTP workload generator

- Review the URL's embedded in `wrk2_lua_scripts/mixed-workload_type_1.lua` to be sure they are correct for your environment.
  The current value of `http://frontend.hotel-res.svc.cluster.local:5000` is valid for a typical "on-cluster" configuration.
- To use an "on-cluster" client, copy the necessary files to `hr-client`, and then log into `hr-client` to continue:
  - `hrclient=$(oc get pod | grep hr-client- | cut -f 1 -d " ")`
  - `oc cp <path-of-repo> hotel-res/"${hrclient}":<path-of-repo>`
    - e.g., `oc cp /root/DeathStarBench hotel-res/"${hrclient}":/root`
  - `oc rsh deployment/hr-client`

### Running HTTP workload generator

##### Template
```bash
cd <path-of-repo>/hotelReservation
./wrk -D exp -t <num-threads> -c <num-conns> -d <duration> -L -s ./wrk2_lua_scripts/mixed-workload_type_1.lua http://frontend.hotel-res.svc.cluster.local:5000 -R <reqs-per-sec>
```

##### Example
```bash
cd /root/DeathStarBench/hotelReservation
./wrk -D exp -t 2 -c 2 -d 30 -L -s ./wrk2_lua_scripts/mixed-workload_type_1.lua http://frontend.hotel-res.svc.cluster.local:5000 -R 2 
```


### View Jaeger traces

Use `oc -n hotel-res get ep | grep jaeger-out` to get the location of jaeger service.

View Jaeger traces by accessing:
- `http://<jaeger-ip-address>:<jaeger-port>`  (off cluster)
- `http://jaeger.hotel-res.svc.cluster.local:6831`  (on cluster)


### Tips

- If you are running on-cluster, you can use the following command to copy files off of the client.
e.g., to copy the results directory from the on-cluster client to the local machine:
  - `hrclient=$(oc get pod | grep hr-client- | cut -f 1 -d " ")`
  - `oc cp hotel-res/${hrclient}:/root/DeathStarBench/hotelReservation/openshift/results /tmp`
