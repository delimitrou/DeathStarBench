##--- Store
# Image store
helm install redis-img bitnami/redis-cluster -n yanqizhang --set global.storageClass=local-storage --set persistence.size=60Gi --set global.redis.password=redisstore --set cluster.nodes=6 --wait
kubectl -n yanqizhang apply -f frontend/config/image_store_redis_cluster.yaml
# Post store
helm install redis-post bitnami/redis-cluster -n yanqizhang --set global.storageClass=local-storage --set persistence.size=60Gi --set global.redis.password=redispost --set cluster.nodes=12 --wait
kubectl -n yanqizhang apply -f post/config/post_store_redis_cluster.yaml
# Social graph store
helm install redis-socialgraph bitnami/redis-cluster -n yanqizhang --set global.storageClass=local-storage --set persistence.size=60Gi --set global.redis.password=redissocialgraph --set cluster.nodes=6 --wait
kubectl -n yanqizhang apply -f ./socialgraph/config/social_store_redis_cluster.yaml
# Timeline store
helm install redis-timeline bitnami/redis-cluster -n yanqizhang --set global.storageClass=local-storage --set persistence.size=60Gi --set global.redis.password=redistimeline --set cluster.nodes=6 --wait
kubectl -n yanqizhang apply -f timeline-update/config/timeline_store_redis_cluster.yaml
##--- pubsub
# Timeline pubsub
helm install redis-tl-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=60Gi --set replica.persistence.size=60Gi --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f timeline-update/config/timeline_pubsub_redis.yaml
# Object detect pubsub
helm install redis-object-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=60Gi --set replica.persistence.size=60Gi --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f object-detect/config/objdet_pubsub_redis.yaml
# Sentiment analysis pubsub
helm install redis-senti-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f ./sentiment/config/senti_pubsub_redis.yaml