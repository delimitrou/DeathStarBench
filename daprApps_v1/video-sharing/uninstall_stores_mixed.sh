##--- Store
# Video store
helm uninstall redis-video -n yanqizhang
kubectl -n yanqizhang delete -f video-frontend/config/video_store_redis_cluster.yaml
# Video store
helm uninstall redis-thumb -n yanqizhang
kubectl -n yanqizhang delete -f video-frontend/config/thumbnail_store_redis_cluster.yaml
# Date store
helm uninstall redis-date -n yanqizhang
kubectl -n yanqizhang delete -f dates/config/date_store_redis.yaml
# Rating store
helm uninstall redis-rating -n yanqizhang
kubectl -n yanqizhang delete -f user-rating/config/rating_store_redis.yaml
# Info store
helm uninstall redis-info -n yanqizhang
kubectl -n yanqizhang delete -f video-info/config/info_store_redis_cluster.yaml
##--- pubsub
# Video pubsub
helm uninstall redis-video-pubsub -n yanqizhang
kubectl -n yanqizhang delete -f video-scale/config/video_pubsub_redis.yaml