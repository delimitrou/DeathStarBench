#!/bin/bash

cd $(dirname $0)/..

function usage()
{
    echo    "Script to update microservice configuration for a specified service/namespace."
    echo    ""
    echo -e "\t-h --help"
    echo -e "\t--namespace='social-network' (the default namespace is 'social-network')"
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
        --namespace)
            NAMESPACE=$VALUE
            ;;
        --micro-service)
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

if [[ -z $NAMESPACE ]]; then
  echo "Using the default namespace: 'social-network'"
  NAMESPACE='social-network'
fi

if [[ -z $MICROSERVICE ]]; then
  echo "You must include the path to the micro-service yaml file"
  usage
  exit 1
fi

oc create -f $MICROSERVICE -n $NAMESPACE --dry-run --save-config -o yaml | oc apply -f - --namespace ${NAMESPACE}

cd - >/dev/null
