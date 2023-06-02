##--- Store
# Image store
helm install redis-img bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false --set replica.persistence.enabled=false --set global.redis.password=redisstore --wait
kubectl -n yanqizhang apply -f frontend/config/image_store_redis.yaml
# Post store
helm install redis-post bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false --set replica.persistence.enabled=false --set global.redis.password=redispost --wait
kubectl -n yanqizhang apply -f post/config/post_store_redis.yaml
# Social graph store
helm install redis-socialgraph bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false  --set replica.persistence.enabled=false --set global.redis.password=redissocialgraph --wait
kubectl -n yanqizhang apply -f ./socialgraph/config/social_store_redis.yaml
# Timeline store
helm install redis-timeline bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false --set replica.persistence.enabled=false --set global.redis.password=redistimeline --wait
kubectl -n yanqizhang apply -f timeline-update/config/timeline_store_redis.yaml
##--- pubsub
# Timeline pubsub
helm install redis-tl-pubsub bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false --set replica.persistence.enabled=false --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f timeline-update/config/timeline_pubsub_redis.yaml
# Object detect pubsub
helm install redis-object-pubsub bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.enabled=false --set replica.persistence.enabled=false --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f object-detect/config/objdet_pubsub_redis.yaml
# Sentiment analysis pubsub
helm install redis-senti-pubsub bitnami/redis -n yanqizhang --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f ./sentiment/config/senti_pubsub_redis.yaml