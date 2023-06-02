#--- Deploy
# Social graph
kubectl -n yanqizhang delete -f socialgraph/k8s/deployment.yaml
# Post
kubectl -n yanqizhang delete -f post/k8s/deployment.yaml
# User
kubectl -n yanqizhang delete -f user/k8s/deployment.yaml
# Timeline update
kubectl -n yanqizhang delete -f timeline-update/k8s/deployment.yaml
# Timeline read
kubectl -n yanqizhang delete -f timeline-read/k8s/deployment.yaml
# Recommend
kubectl -n yanqizhang delete -f recommend/k8s/deployment.yaml
# Object detect
kubectl -n yanqizhang delete -f object-detect/k8s/deployment.yaml
# Sentiment
kubectl -n yanqizhang delete -f sentiment/k8s/deployment.yaml
# Translate
kubectl -n yanqizhang delete -f translate/k8s/deployment.yaml
# Frontend
kubectl -n yanqizhang delete -f frontend/k8s/deployment.yaml
# Proxy
kubectl -n yanqizhang delete -f proxy/k8s/deployment.yaml
##--- Service
# Frontend
kubectl -n yanqizhang delete -f frontend/k8s/service_frontend.yaml
# Social graph
kubectl -n yanqizhang delete -f socialgraph/k8s/service_social.yaml
# User
kubectl -n yanqizhang delete -f user/k8s/service_user.yaml
# Recommend
kubectl -n yanqizhang delete -f recommend/k8s/service_recmd.yaml
# Translate
kubectl -n yanqizhang delete -f translate/k8s/service_transl.yaml
# Proxy
kubectl -n yanqizhang delete -f proxy/k8s/service_proxy.yaml