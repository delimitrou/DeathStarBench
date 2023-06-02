##--- Store
# Video store
helm install redis-video-store bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redisstore --wait
kubectl -n yanqizhang apply -f video-frontend/config/video_store_redis.yaml
# Date store
helm install redis-date bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redisdate --wait
kubectl -n yanqizhang apply -f dates/config/date_store_redis.yaml
# Rating store
helm install redis-rating bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi  --set replica.persistence.size=10Gi --set global.redis.password=redisrating --wait
kubectl -n yanqizhang apply -f user-rating/config/rating_store_redis.yaml
# Info store
helm install redis-info bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redisinfo --wait
kubectl -n yanqizhang apply -f video-info/config/info_store_redis.yaml
##--- pubsub
# Video pubsub
helm install redis-video-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f video-scale/config/video_pubsub_redis.yaml
#--- Deploy
# Video-info
kubectl -n yanqizhang apply -f video-info/k8s/deployment.yaml
# User-rating
kubectl -n yanqizhang apply -f user-rating/k8s/deployment.yaml
# Trending
kubectl -n yanqizhang apply -f trending/k8s/deployment.yaml
# Dates
kubectl -n yanqizhang apply -f dates/k8s/deployment.yaml
# Video-scale
kubectl -n yanqizhang apply -f video-scale/k8s/deployment.yaml
# Video-thumbnail
kubectl -n yanqizhang apply -f video-thumbnail/k8s/deployment.yaml
# Video-frontend
kubectl -n yanqizhang apply -f video-frontend/k8s/deployment.yaml
##--- Service
# Video-frontend
kubectl -n yanqizhang apply -f video-frontend/k8s/service_frontend.yaml
# Trending
kubectl -n yanqizhang apply -f trending/k8s/service_trending.yaml