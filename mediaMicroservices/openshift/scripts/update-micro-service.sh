#!/bin/bash

cd $(dirname $0)/..

NAMESPACE='media-microsvc'

function usage()
{
    echo    "Script to generate destination rule of all services from a given namespace"
    echo    ""
    echo -e "\t-h --help"
    echo -e "\t--namespace='media-microsvc' (the default namespace is 'media-microsvc')"
    echo -e "\t--micro-service='path to micro-service yaml file'"
    echo -e "\tPlease, do not include space between the parameter and value"
    echo    ""
}

while [ "$1" != "" ]; do
    PARAM=`echo $1 | awk -F= '{print $1}'`
    VALUE=`echo $1 | awk -F= '{print $2}'`
    case $PARAM in
        -h | --help)
            usage
            exit
            ;;
        -n | --namespace)
            NAMESPACE=$VALUE
            ;;
        -m | --micro-service)
            MICROSERVICE=$VALUE
            ;;
        *)
            echo "ERROR: unknown parameter \"$PARAM\""
            usage
            exit 1
            ;;
    esac
    shift
done

if [[ -z $MICROSERVICE ]]; then
  echo "You must include the path to the micro-service yaml file"
  usage
  exit
fi

oc create -f $MICROSERVICE --namespace ${NAMESPACE} --dry-run --save-config -o yaml | oc apply -f - --namespace ${NAMESPACE}

cd -
