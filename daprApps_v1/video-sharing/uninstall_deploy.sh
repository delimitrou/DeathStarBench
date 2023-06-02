#--- Deploy
# Video-info
kubectl -n yanqizhang delete -f video-info/k8s/deployment.yaml
# User-rating
kubectl -n yanqizhang delete -f user-rating/k8s/deployment.yaml
# Trending
kubectl -n yanqizhang delete -f trending/k8s/deployment.yaml
# Dates
kubectl -n yanqizhang delete -f dates/k8s/deployment.yaml
# Video-scale
kubectl -n yanqizhang delete -f video-scale/k8s/deployment.yaml
# Video-thumbnail
kubectl -n yanqizhang delete -f video-thumbnail/k8s/deployment.yaml
# Video-frontend
kubectl -n yanqizhang delete -f video-frontend/k8s/deployment.yaml
##--- Service
# Video-frontend
kubectl -n yanqizhang delete -f video-frontend/k8s/service_frontend.yaml
# Trending
kubectl -n yanqizhang delete -f trending/k8s/service_trending.yaml