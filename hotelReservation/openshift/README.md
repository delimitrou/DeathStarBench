# Hotel Reservations on OpenShift 4.x

## Pre-requirements

- A running OpenShift cluster is needed.
- The user should be authenticated to this cluster e.g., `oc login`
- The cluster should have a namespace `hotel-res`, if not create using `oc create namespace hotel-res`
- Pre-requirements mentioned [here](https://github.com/delimitrou/DeathStarBench/blob/master/hotelReservation/README.md) should be met.

## Running the Hotel Reservation application

### Before you start

- Ensure that the necessary local images have been made. Before executing this script make sure already existing images related to this application are deleted both on cluster and your local environment using podman.
  - `<path-of-repo>/hotelReservation/openshift/scripts/build-docker-img.sh`
- If necessary, update the addresses in `<path-of-repo>/hotelReservation/openshift/configmaps/config.json`
  - Currently the addresses are in a fairly generic form supported by on-cluster DNS. As long as
    access is from within the cluster and the name of the namespace has not been changed, it should be OK.

### Deploy services

run `<path-of-repo>/hotelReservation/openshift/scripts/deploy.sh`
and wait for `oc -n hotel-res get pod` to show all pods with status `Running`.


### Prepare HTTP workload generator
- To use an "on-cluster" client, copy the necessary files to `hr-client`, and then log into `hr-client` to continue:
  - `hrclient=$(oc get pod | grep hr-client- | cut -f 1 -d " ")`
  - `oc cp <path-of-repo-in-local> hotel-res/"${hrclient}":<path-of-repo>`
    - e.g., `oc cp /root/DeathStarBench hotel-res/"${hrclient}":/root`
  - `oc rsh deployment/hr-client`

### Running HTTP workload generator

##### Template
```bash
cd <path-of-repo>/hotelReservation/wrk2
make clean
make
./wrk -D exp -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/hotel-reservation/mixed-workload_type_1.lua http://frontend.hotel-res.svc.cluster.local:5000 -R <reqs-per-sec>
```

##### Example
```bash
cd /root/DeathStarBench/hotelReservation/wrk2
make clean
make
./wrk -D exp -t 2 -c 2 -d 30 -L -s ./scripts/hotel-reservation/mixed-workload_type_1.lua http://frontend.hotel-res.svc.cluster.local:5000 -R 2
```


### View Jaeger traces

Use `oc -n hotel-res get ep | grep jaeger-out` to get the location of jaeger service.

View Jaeger traces by accessing:
- `http://<jaeger-ip-address>:<jaeger-port>`  (off cluster)
- `oc expose service jaeger-out` (on cluster)
- `oc get route` (on cluster)
- `http://<jeager-route-url>`  (on cluster)


### Tips

- If you are running on-cluster, you can use the following command to copy files off of the client.
e.g., to copy the results directory from the on-cluster client to the local machine:
  - `hrclient=$(oc get pod | grep hr-client- | cut -f 1 -d " ")`
  - `oc cp hotel-res/${hrclient}:/root/DeathStarBench/hotelReservation/openshift/results /tmp`
