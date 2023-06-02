##--- Store
# Image store
helm uninstall redis-img -n yanqizhang
kubectl -n yanqizhang delete -f frontend/config/image_store_redis_cluster.yaml
# Post store
helm uninstall redis-post -n yanqizhang
kubectl -n yanqizhang delete -f post/config/post_store_redis_cluster.yaml
# Social graph store
helm uninstall redis-socialgraph -n yanqizhang
kubectl -n yanqizhang delete -f ./socialgraph/config/social_store_redis_cluster.yaml
# Timeline store
helm uninstall redis-timeline -n yanqizhang
kubectl -n yanqizhang delete -f timeline-update/config/timeline_store_redis_cluster.yaml
##--- pubsub
# Timeline pubsub
helm uninstall redis-tl-pubsub -n yanqizhang
kubectl -n yanqizhang delete -f timeline-update/config/timeline_pubsub_redis.yaml
# Object detect pubsub
helm uninstall redis-object-pubsub -n yanqizhang
kubectl -n yanqizhang delete -f object-detect/config/objdet_pubsub_redis.yaml
# Sentiment analysis pubsub
helm uninstall redis-senti-pubsub -n yanqizhang
kubectl -n yanqizhang delete -f ./sentiment/config/senti_pubsub_redis.yaml
# delete pvc
kubectl -n yanqizhang delete pvc --all