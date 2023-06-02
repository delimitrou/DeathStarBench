# ns=yanqizhang-profile
ns=$1
##--- Store
# Image store
helm install vipipe-redis-image bitnami/redis -n $ns -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi --set replica.persistence.size=20Gi --set global.redis.password=redisstore --wait
kubectl -n $ns apply -f config/image_store_redis.yaml
# Video store
helm install vipipe-redis-video bitnami/redis -n $ns -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi --set replica.persistence.size=20Gi --set global.redis.password=redisstore --wait
kubectl -n $ns apply -f config/video_store_redis.yaml
##--- pubsub
# vipipe events
helm install redis-vpipe-pubsub bitnami/redis -n $ns -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi --set replica.persistence.size=20Gi --set global.redis.password=redispubsub --wait
kubectl -n $ns apply -f config/vpipe_pubsub_redis.yaml