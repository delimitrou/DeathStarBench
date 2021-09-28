# Hotel Reservation

The application implements a hotel reservation service, build with Go and gRPC, and starting from the open-source project https://github.com/harlow/go-micro-services. The initial project is extended in several ways, including adding back-end in-memory and persistent databases, adding a recommender system for obtaining hotel recommendations, and adding the functionality to place a hotel reservation. 

<!-- ## Application Structure -->

<!-- ![Social Network Architecture](socialNet_arch.png) -->

Supported actions: 
* Get profile and rates of nearby hotels available during given time periods
* Recommend hotels based on user provided metrics
* Place reservations

## Pre-requirements
- Docker
- Docker-compose
- luarocks (apt-get install luarocks)
- luasocket (luarocks install luasocket)

## Running the social network application
### Before you start
- Install Docker and Docker Compose.
- Make sure exposed ports in docker-compose files are available
- Consider which platform you want to use (docker-compose/openshift/kubernetes)
    - Build the required images using the proper method
        - In case of docker-compose => docker-compose build
        - In case of Openshift => run the build script according to the readme.
        - In case of kubernetes => run the build script according to the readme.

### Running the containers
##### Docker-compose
- NOTLS: Start docker containers by running `docker-compose up -d`. All images will be pulled from Docker Hub.
- TLS: Start docker containers by running `TLS=1 docker-compose up -d`. All the gRPC communications will be protected by TLS.
- TLS with spcified ciphersuite: Start docker containers by running `TLS=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 docker-compose up -d`. The available cipher suite can be find at the file [options.go](tls/options.go#L21).

Check if TLS is enabled or not: `docker-compose logs <service> | grep TLS`.

##### Openshift
Read the Readme file in Openshift directory.

##### Kubernetes
Read the Readme file in Kubernetes directory.

#### workload generation
```bash
./wrk2/wrk -D exp -t <num-threads> -c <num-conns> -d <duration> -L -s ./wrk2/scripts/hotel-reservation/mixed-workload_type_1.lua http://x.x.x.x:5000 -R <reqs-per-sec>
```

### Questions and contact

You are welcome to submit a pull request if you find a bug or have extended the application in an interesting way. For any questions please contact us at: <microservices-bench-L@list.cornell.edu>
