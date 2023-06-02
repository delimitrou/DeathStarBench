# ns=yanqizhang-profile
ns=$1
#--- Deploy
# Video-meta
kubectl -n $ns apply -f video-meta/k8s/deployment.yaml
# Video-scene
kubectl -n $ns apply -f video-scene/k8s/deployment.yaml
# Face-detect
kubectl -n $ns apply -f face-detect/k8s/deployment.yaml
##--- Service
# Meta service
kubectl -n $ns apply -f video-meta/k8s/meta_service.yaml