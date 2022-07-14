# Media Microservices on OpenShift 4.x

## Pre-requirements

- A running OpenShift cluster is needed.
- The user should be authenticated to this cluster e.g., `oc login`
- Pre-requirements mentioned [here](https://github.com/delimitrou/DeathStarBench/blob/master/mediaMicroservices/README.md) should be met.

## Running the media service application

### Before you start

Set the resolver to be the ip address for dns service of your cluster in `<path-of-repo>/mediaMicroservices/openshift/configmaps/nginx.conf`.
This command should show the ip address: `oc describe dns.operator/default`

### Deploy services

run `<path-of-repo>/mediaMicroservices/openshift/scripts/deploy-all-services-and-configurations.sh`
and wait `oc -n media-microsvc get pod` to show all pods with status `Running`.

### Register users and movie information

- If you are using an "off-cluster" client then use `oc -n media-microsvc describe nginx-web-server` or `oc -n media-microsvc get ep | grep nginx-web-server` to get the ip address for the web server, and then then update the address and port number (:8080) as needed in these files. In case of routes get the route url using `oc get route -n media-microsvc`:
  - `sed -i 's/127.0.0.1:8080/<your-server-endpoint-address:port>/g' <path-of-repo>/mediaMicroservices/scripts/write_movie_info.py`
  - `sed -i 's/127.0.0.1:8080/<your-server-endpoint-address:port>/g' <path-of-repo>/mediaMicroservices/scripts/register_users.sh`
  - `sed -i 's/127.0.0.1:8080/<your-server-endpoint-address:port>/g' <path-of-repo>/mediaMicroservices/scripts/register_movies.sh`
  - `sed -i 's/127.0.0.1:8080/<your-server-endpoint-address:port>/g' <path-of-repo>/mediaMicroservices/scripts/compose_review.sh`
- If you are using an "on-cluster" client, then the address of the web server is `nginx-web-server.media-microsvc.svc.cluster.local:8080`
- If you are running "on-cluster" copy necessary files to mms-client, and then log into mms-client to continue:
  - `mmsclient=$(oc get pod | grep mms-client- | cut -f 1 -d " ")`
  - `oc cp <path-of-repo> media-microsvc/"${mmsclient}":<path-of-repo>`
    - e.g., `oc cp /<your-repo-path>/DeathStarBench media-microsvc/"${mmsclient}":/root`
  - `oc rsh deployment/mms-client`
- For both on and off cluster clients, initialize the databases:
  - `cd /root/DeathStarBench/mediaMicroservices/scripts`
  - `python3 <path-of-repo>/mediaMicroservices/scripts/write_movie_info.py && <path-of-repo>/mediaMicroservices/scripts/register_users.sh`
    - e.g., `python3 /root/DeathStarBench/mediaMicroservices/scripts/write_movie_info.py && /root/DeathStarBench/mediaMicroservices/scripts/register_users.sh`

### Running HTTP workload generator

#### Compose reviews

```bash
cd <path-of-repo>/mediaMicroservices/wrk2
make clean
make
./wrk -D exp -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/media-microservices/compose-review.lua http://<webserver-address>:8080/wrk2-api/review/compose -R <reqs-per-sec>
#   e.g., ./wrk -D exp -t 2 -c 2 -d 30 -L -s ./scripts/media-microservices/compose-review.lua http://nginx-web-server.media-microsvc.svc.cluster.local:8080/wrk2-api/review/compose -R 2
```

### View Jaeger traces

Use `oc -n media-microsvc get ep | grep jaeger-out` to get the location of jaeger service.

View Jaeger traces by accessing `http://<jaeger-ip-address>:<jaeger-port>` 


### Tips

If you are running on-cluster, you can use the following command to copy files off of the client.
e.g., to copy the results directory from the on-cluster client to the local machine:
  - `mmsclient=$(oc get pod | grep mms-client- | cut -f 1 -d " ")`
  - `oc cp media-microsvc/${mmsclient}:/root/DeathStarBench/mediaMicroservices/openshift/results /tmp`
