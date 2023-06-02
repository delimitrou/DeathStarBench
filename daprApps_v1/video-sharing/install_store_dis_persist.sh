##--- Store
# Video store
helm install redis-video-store bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false --set replica.persistence.enabled=false --set global.redis.password=redisstore --wait
kubectl -n yanqizhang apply -f video-frontend/config/video_store_redis.yaml
# Thumbnail store
helm install redis-thumb bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redisthumb --wait
kubectl -n yanqizhang apply -f video-frontend/config/thumbnail_store_redis.yaml
# Date store
helm install redis-date bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false --set replica.persistence.enabled=false --set global.redis.password=redisdate --wait
kubectl -n yanqizhang apply -f dates/config/date_store_redis.yaml
# Rating store
helm install redis-rating bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false  --set replica.persistence.enabled=false --set global.redis.password=redisrating --wait
kubectl -n yanqizhang apply -f user-rating/config/rating_store_redis.yaml
# Info store
helm install redis-info bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false --set replica.persistence.enabled=false --set global.redis.password=redisinfo --wait
kubectl -n yanqizhang apply -f video-info/config/info_store_redis.yaml
##--- pubsub
# Video pubsub
helm install redis-video-pubsub bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false --set replica.persistence.enabled=false --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f video-scale/config/video_pubsub_redis.yaml