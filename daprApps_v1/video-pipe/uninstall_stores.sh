# ns=yanqizhang-profile
ns=$1
##--- Store
# Image store
helm uninstall vipipe-redis-image -n $ns
kubectl -n $ns delete -f config/image_store_redis.yaml
# Video store
helm uninstall vipipe-redis-video -n $ns
kubectl -n $ns delete -f config/video_store_redis.yaml
##--- pubsub
# Events pubsub
helm uninstall redis-vpipe-pubsub -n $ns
kubectl -n $ns delete -f config/vpipe_pubsub_redis.yaml