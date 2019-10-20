# Media Microservices on Kubernetes

## Pre-requirements

- A running kubernetes cluster is needed.
- This repo should be put in the same path on every node of the cluster for volume mount (here we use `/root/DeathStarBench/`).
- Pre-requirements mentioned [here](https://github.com/delimitrou/DeathStarBench/blob/master/mediaMicroservices/README.md) should be met.

## Running the media service application

### Before you start

Set the resolver to be the FQDN of the core-dns or kube-dns  service of your cluster in `<path-of-repo>/mediaMicroservices/nginx-web-server/conf/nginx-k8s.conf`. If you do not deploy core-dns or kube-dns, I am not sure whether all things below still work.

### Deploy services

Run `kubectl apply -f <path-of-repo>/mediaMicroservices/k8s-yaml/` and wait `kubectl -n media-microsvc get pod` to show all pods with status `Running`.

### Register users and movie information

- Use `kubectl -n media-microsvc get svc nginx-web-server ` to get its cluster-ip.
- Paste the cluster ip at `<path-of-repo>/mediaMicroservices/scripts/write_movie_info.py:102 & 109` and `<path-of-repo>/mediaMicroservices/scripts/register_users.sh:5`
- `python3 <path-of-repo>/mediaMicroservices/scripts/write_movie_info.py && <path-of-repo>/mediaMicroservices/scripts/register_users.sh`

### Running HTTP workload generator

#### Compose reviews

Paste the cluster ip at `<path-of-repo>/mediaMicroservices/wrk2/scripts/media-microservices/compose-review.lua:1032`

Then

```bash
cd <path-of-repo>/mediaMicroservices/wrk2
./wrk -D exp -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/media-microservices/compose-review.lua http://<cluster-ip>:8080/wrk2-api/review/compose -R <reqs-per-sec>
```

#### View Jaeger traces

Use `kubectl -n media-microsvc get svc jaeger-out` to get the NodePort of jaeger service.

 View Jaeger traces by accessing `http://<a-node-ip>:<NodePort>` 