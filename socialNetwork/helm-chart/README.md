# Social Network Microservices Helm Chart #

# What is Helm Chart ##
Helm charts are packages containing Kubernetes yaml files. Its main goal is to automate the deployment of an application on a Kubernetes cluster. It allows for defining the behaviour of an application and an easy way of manipulating application's parameters. Packages are easily portable across platforms.

## Purpose of this project ##
The main goal of this project is to automate the process of deploying Social Network Microservices on a Kubernetes cluster natively using helm chart. 

## Structure of helm chart  ##
Every microservice is packaged into its own isolated helm chart. All these packages are assembled under one main helm chart. Microservices share the same deployment, service and configmap files templates which are parameterized using values from `values.yaml` file in each microsevice package. Helm charts also share the same config files. The main helm chart contains global values which are shared among microservices but can be individually overridden.

## Shared config files ##
Microservices with same purposes share config files which can be found under `templates/configs` in the main helm chart.
The following subdirectories reflect different types of microservices and contain respective config files.

```
templates/
    configs/
        media-service/
        mongo/
        nginx/
        redis/
        other/
```

All mongo, nginx and redis pods will use the same respective config file from the template. The media-frontend service will use the config files in the media-service folder. The rest of the microservices (that use the all-purpose social network microservices image) will use the config file located in the "other" folder.

In order to override a given value for a config file, a new config (with the new content) must be placed under the same filename in the given microservice template.

## Custom images ##
`nginx-thrift` and `media-frontend` services require mounted lua-scripts (and other files). For this reason, in both of these pods, there is an init container added that pulls the rquired files from the public DSB repository and mounts them in the right path before the container is started.

### Changes to config files ###
In nginx config file, both in `nginx-thrift` and `media-frontend` pods, resolver is set to a value retrieved from global values under `resolverName`. It is set to `kube-dns.kube-system.svc.cluster.local`. <br />

Also `fqdn_suffix` is set to be an environemt variable. Its value is set in `values.yaml` file in corresponding pods. 


### Changes to lua scripts ###
In every lua script where there is a call to a service, a Kubernetes suffix must be added to the name of the service. The suffix is read from the environment variable `fqdn_suffix`. It can be achieved by adding the following code in lua scripts:
```
local k8s_suffix = os.getenv("fqdn_suffix")
service_name = service_name .. k8s_suffix 
```

`fqdn_suffix` is in form `NAMESPACE.svc.cluster.local`. <br />
By default, Kuberentes’ resolver is configured with search domains. However, when we use custom nginx resolver, we need to specify FQDN (Full Qualified Domain Name).


## Deployment ##
In order to deploy services using helm chart, helm needs to be installed (https://helm.sh/).
The following line shows a default deployment of Social Nework Microservices using helm chart:

```
helm install RELEASE_NAME HELM_CHART_REPO_PATH
```

### Setting namespace ###

```
helm install RELEASE_NAME HELM_CHAHELM_CHART_REPO_PATHRT_PATH -n NAMESPACE
```

### Overriding default values ###

#### Default replicas ####
```
helm install RELEASE_NAME HELM_CHART_REPO_PATH --set global.replicas=2
```

#### Default image pull policy ####
```
helm install RELEASE_NAME HELM_CHART_REPO_PATH --set-string global.imagePullPolicy=Always
```

#### compose-post-service pod replicas count ####
```
helm install RELEASE_NAME HELM_CHART_REPO_PATH --set compose-post-service.replicas=3
```

### Setting topology spread constraints ###
Kubernetes allows for controlling pods spread accross the cluster. We can specify the same spread constraints for all the pods or for the given pod.

(https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/)

#### Same spread constrains for all the pods ####
```
helm install RELEASE_NAME HELM_CHART_REPO_PATH \
 --set-string global.topologySpreadConstraints="- maxSkew: 1
    topologyKey: node
    whenUnsatisfiable: DoNotSchedule
    labelSelector:
      matchLabels:
        service: {{ .Values.name }}"
```

#### Unique spread constrains for compose-post-service ####
```
helm install RELEASE_NAME HELM_CHART_REPO_PATH \
 --set-string compose-post-service.topologySpreadConstraints="- maxSkew: 1
    topologyKey: node
    whenUnsatisfiable: DoNotSchedule
    labelSelector:
      matchLabels:
        service: {{ .Values.name }}"
```

Note: indentations are important in strings that are being passed as parameters.


### Setting resources ###
Kubernetes allows resources (CPU, memory) to be set and assigned to a container. By default, in the helm chart, none of the containers in any service has any resource constraints. We can specify the same resource constraints for all the containers or for the given one.
Example:

#### Same resources for all the containers ####
```
helm install RELEASE_NAME HELM_CHART_REPO_PATH \
  --set-string global.resources="requests: 
      memory: "64Mi"
      cpu: "250m"
    limits:
      memory: "128Mi"
      cpu: "2""
```

#### Setting resources for the compose-post-service container ####

```
helm install RELEASE_NAME HELM_CHART_REPO_PATH \
   --set-string compose-post-service.container.resources="requests: 
      memory: "64Mi"
      cpu: "250m"
    limits:
      memory: "128Mi"
      cpu: "2""
```

#### Same resources for all the containers and unique for compose-post-service ####
```
helm install RELEASE_NAME HELM_CHART_REPO_PATH \
  --set-string global.resources="requests: 
      memory: "64Mi"
      cpu: "250m"
    limits:
      memory: "128Mi"
      cpu: "2"" \
  --set-string compose-post-service.container.resources="requests: 
      memory: "64Mi"
      cpu: "500m"
    limits:
      memory: "128Mi"
      cpu: "4""
```
Note: indentations are important in strings that are being passed as parameters.

If an entry under resources is not specified the value is retrieved from global values (which can be overriden during deployment). <br />
(Documnetation on Kubernetes container resources management: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

## Adding new service ##
To add new service to the helm chart the following steps need to be followed:

- add new helm chart under `charts/`
- add new service to dependencies in `Chart.yaml` file in the main helm chart
```
- name: SERVICE_NAME
    version: VERSION
    repository: file::/charts/SERVICE_NAME
```

- add new service in `templates/configs/other/service-config.tpl`
