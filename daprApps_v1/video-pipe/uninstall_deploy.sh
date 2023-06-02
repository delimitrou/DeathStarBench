ns=$1
#--- Deploy
# Video-meta
kubectl -n $ns delete -f video-meta/k8s/deployment.yaml
# Video-scene
kubectl -n $ns delete -f video-scene/k8s/deployment.yaml
# Face-detect
kubectl -n $ns delete -f face-detect/k8s/deployment.yaml
##--- Service
# Meta service
kubectl -n $ns delete -f video-meta/k8s/meta_service.yaml