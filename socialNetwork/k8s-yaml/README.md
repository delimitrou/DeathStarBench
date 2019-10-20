# Social Network Microservices on Kubernetes

## Pre-requirements

- A running kubernetes cluster is needed.
- This repo should be put in the same path on every node of the cluster for volume mount (here we use `/root/DeathStarBench/`).
- Pre-requirements mentioned [here](https://github.com/delimitrou/DeathStarBench/blob/master/socialNetwork/README.md) should be met.

## Running the social network application on kubernetes

### Before you start

Set the resolver to be the FQDN of the core-dns or kube-dns  service of your cluster in `<path-of-repo>/socialNetwork/nginx-web-server/conf/nginx-k8s.conf` and `<path-of-repo>/socialNetwork/media-frontend/conf/nginx-k8s.conf`. If you do not deploy core-dns or kube-dns, I am not sure whether all things below still work.

### Deploy services

Run `kubectl apply -f <path-of-repo>/socialNetwork/k8s-yaml/` and wait `kubectl -n social-network get pod` to show all pods with status `Running`.

### Register users and construct social graphs

- Use `kubectl -n social-network get svc nginx-thrift ` to get its cluster-ip.
- Paste the cluster ip at `<path-of-repo>/socialNetwork/scripts/init_social_graph.py:72`
- Register users and construct social graph by running `python3 <path-of-repo>/socialNetwork/scripts/init_social_graph.py`. This will initialize a social graph based on [Reed98 Facebook Networks](http://networkrepository.com/socfb-Reed98.php), with 962 users and 18.8K social graph edges. 

### Running HTTP workload generator

#### Compose posts

Paste the cluster ip at `<path-of-repo>/socialNetwork/wrk2/scripts/social-network/compose-post.lua:66`

Then

```bash
cd <path-of-repo>/socialNetwork/wrk2
./wrk -D exp -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/social-network/compose-post.lua http://<clulster-ip>:8080/wrk2-api/post/compose -R <reqs-per-sec>
```

#### Read home timelines

Paste the cluster ip at `<path-of-repo>/socialNetwork/wrk2/scripts/social-network/read-home-timeline.lua:16`

Then

```bash
cd <path-of-repo>/socialNetwork/wrk2
./wrk -D exp -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/social-network/read-home-timeline.lua http://<clulster-ip>:8080/wrk2-api/home-timeline/read -R <reqs-per-sec>
```

#### Read user timelines

Paste the cluster ip at `<path-of-repo>/socialNetwork/wrk2/scripts/social-network/read-user-timeline.lua:16`

Then

```bash
cd <path-of-repo>/socialNetwork/wrk2
./wrk -D exp -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/social-network/read-user-timeline.lua http://<clulster-ip>:8080/wrk2-api/user-timeline/read -R <reqs-per-sec>
```

#### View Jaeger traces

Use `kubectl -n social-network get svc jaeger-out` to get the NodePort of jaeger service.

 View Jaeger traces by accessing `http://<a-node-ip>:<NodePort>` 

## TODO

Obviously the use of cluster-ip should be more configurable, pasting it in the code is kind of annoying. I am not familiar with lua, and I just leave it there.  You are really welcome to submit a pull request!

## Questions and contact

If you have questions or advices about the k8s-yamls, contact me at myles.l.pan@gmail.com