### How to Run

To build a docker container for a SERVICE service, run:
```
.\cd chatReactive
.\sbt SERVICE/docker:publishLocal 
```
(Example: .\sbt microservice_1/docker:publishLocal)


To run the cluster, run `docker-compose up` after building all of the necessary containers.
You can also run all above steps with `.\scripts\build-all`

Heavily based on 
[akka-sample-cluster-docker-compose-scala](https://github.com/akka/akka-sample-cluster-docker-compose-scala)
===========================
