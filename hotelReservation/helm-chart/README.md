# Hotel Reservation Microservices Helm Chart #
  
# What is Helm Chart ##
Helm charts are packages containing Kubernetes yaml files. Its main goal is to automate the deployment of an application on a Kubernetes cluster. It allows for defining the behaviour of an application and an easy way of manipulating application's parameters. Packages are easily portable across platforms.

## Purpose of this project ##
The main goal of this project is to automate the process of deploying Hotel Reservation Microservices on a Kubernetes cluster natively using helm chart.

## Structure of helm chart  ##
Every microservice is packaged into its own isolated helm chart. All these packages are assembled under one main helm chart. Microservices share the same deployment, service and configmap files templates which are parameterized using values from `values.yaml` file in each microsevice package. Helm charts also share the same config files. The main helm chart contains global values which are shared among microservices but can be individually overridden.

## Service config file ##
All Hotel Reservation services (that use the all-purpose hotel reservation microservices image) will use the config file found under `templates/configs` in the main helm chart.
```
templates/
    configs/
       service-config.tpl 
```

## Deployment ##
In order to deploy services using helm chart, helm needs to be installed (https://helm.sh/).
The following line shows a default deployment of Hotel Reservation Microservices using helm chart:

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

#### frontend pod replicas count ####
```
helm install RELEASE_NAME HELM_CHART_REPO_PATH --set frontend.replicas=3
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

#### Unique spread constrains for frontend ####
```
helm install RELEASE_NAME HELM_CHART_REPO_PATH \
 --set-string frontend.topologySpreadConstraints="- maxSkew: 1
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

#### Setting resources for the frontend container ####

```
helm install RELEASE_NAME HELM_CHART_REPO_PATH \
   --set-string frontend.container.resources="requests:
      memory: "64Mi"
      cpu: "250m"
    limits:
      memory: "128Mi"
      cpu: "2""
```

#### Same resources for all the containers and unique for frontend ####
```
helm install RELEASE_NAME HELM_CHART_REPO_PATH \
  --set-string global.resources="requests:
      memory: "64Mi"
      cpu: "250m"
    limits:
      memory: "128Mi"
      cpu: "2"" \
  --set-string frontend.container.resources="requests:
      memory: "64Mi"
      cpu: "500m"
    limits:
      memory: "128Mi"
      cpu: "4""
```
Note: indentations are important in strings that are being passed as parameters.

If an entry under resources is not specified the value is retrieved from global values (which can be overriden during deployment). <br />
(Documnetation on Kubernetes container resources management: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
