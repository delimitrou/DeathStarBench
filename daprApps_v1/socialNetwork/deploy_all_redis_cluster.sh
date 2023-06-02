##--- Store
# Image store
helm install redis-img bitnami/redis-cluster -n yanqizhang --set global.storageClass=local-storage --set persistence.size=20Gi --set global.redis.password=redisstore --set cluster.nodes=6 --wait
kubectl -n yanqizhang apply -f frontend/config/image_store_redis_cluster.yaml
# Post store
helm install redis-post bitnami/redis-cluster -n yanqizhang --set global.storageClass=local-storage --set persistence.size=20Gi --set global.redis.password=redispost --set cluster.nodes=6 --wait
kubectl -n yanqizhang apply -f post/config/post_store_redis_cluster.yaml
# Social graph store
helm install redis-socialgraph bitnami/redis-cluster -n yanqizhang --set global.storageClass=local-storage --set persistence.size=20Gi --set global.redis.password=redissocialgraph --set cluster.nodes=6 --wait
kubectl -n yanqizhang apply -f ./socialgraph/config/social_store_redis_cluster.yaml
# Timeline store
helm install redis-timeline bitnami/redis-cluster -n yanqizhang --set global.storageClass=local-storage --set persistence.size=20Gi --set global.redis.password=redistimeline --set cluster.nodes=6 --wait
kubectl -n yanqizhang apply -f timeline-update/config/timeline_store_redis_cluster.yaml
##--- pubsub
# Timeline pubsub
helm install redis-tl-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi --set replica.persistence.size=20Gi --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f timeline-update/config/timeline_pubsub_redis.yaml
# Object detect pubsub
helm install redis-object-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi --set replica.persistence.size=20Gi --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f object-detect/config/objdet_pubsub_redis.yaml
# Sentiment analysis pubsub
helm install redis-senti-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redispubsub --wait
kubectl -n yanqizhang apply -f ./sentiment/config/senti_pubsub_redis.yaml
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