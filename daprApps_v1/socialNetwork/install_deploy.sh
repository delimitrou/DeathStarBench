#--- Deploy
# Social graph
kubectl -n yanqizhang apply -f socialgraph/k8s/deployment.yaml
# Post
kubectl -n yanqizhang apply -f post/k8s/deployment.yaml
# User
kubectl -n yanqizhang apply -f user/k8s/deployment.yaml
# Timeline update
kubectl -n yanqizhang apply -f timeline-update/k8s/deployment.yaml
# Timeline read
kubectl -n yanqizhang apply -f timeline-read/k8s/deployment.yaml
# Recommend
kubectl -n yanqizhang apply -f recommend/k8s/deployment.yaml
# Object detect
kubectl -n yanqizhang apply -f object-detect/k8s/deployment.yaml
# Sentiment
kubectl -n yanqizhang apply -f sentiment/k8s/deployment.yaml
# Translate
kubectl -n yanqizhang apply -f translate/k8s/deployment.yaml
# Frontend
kubectl -n yanqizhang apply -f frontend/k8s/deployment.yaml
# Proxy
kubectl -n yanqizhang apply -f proxy/k8s/deployment.yaml
##--- Service
# Frontend
kubectl -n yanqizhang apply -f frontend/k8s/service_frontend.yaml
# Social graph
kubectl -n yanqizhang apply -f socialgraph/k8s/service_social.yaml
# User
kubectl -n yanqizhang apply -f user/k8s/service_user.yaml
# Recommend
kubectl -n yanqizhang apply -f recommend/k8s/service_recmd.yaml
# Translate
kubectl -n yanqizhang apply -f translate/k8s/service_transl.yaml
# Proxy
kubectl -n yanqizhang apply -f proxy/k8s/service_proxy.yaml