# Deploy benchnmark

## State stores
- Image store
    Deploy redis as image store
    ```bash
    helm install redis-img bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi --set replica.persistence.size=20Gi --set global.redis.password=redisstore --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f frontend/config/image_store_redis.yaml
    ```

- Post store
    Deploy redis as post store
    ```bash
    helm install redis-post bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi --set replica.persistence.size=20Gi --set global.redis.password=redispost --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f post/config/post_store_redis.yaml
    ```

- Social graph store
    Deploy redis as post store
    ```bash
    helm install redis-socialgraph bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi  --set replica.persistence.size=20Gi --set global.redis.password=redissocialgraph --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f ./socialgraph/config/social_store_redis.yaml
    ```

- Timeline store
    Deploy redis as post store
    ```bash
    helm install redis-timeline bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi --set replica.persistence.size=20Gi --set global.redis.password=redistimeline --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f timeline-update/config/timeline_store_redis.yaml
    ```

## Pubsub

- Timeline pubsub
    Deploy redis as pubsub
    ```bash
    helm install redis-tl-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi --set replica.persistence.size=20Gi --set global.redis.password=redispubsub --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f timeline-update/config/timeline_pubsub_redis.yaml
    ```

- Object detect pubsub
    Deploy redis as pubsub
    ```bash
    helm install redis-object-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=20Gi --set replica.persistence.size=20Gi --set global.redis.password=redispubsub --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f object-detect/config/objdet_pubsub_redis.yaml
    ```

- Sentiment analysis pubsub
    Deploy redis as pubsub
    ```bash
    helm install redis-senti-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redispubsub --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f ./sentiment/config/senti_pubsub_redis.yaml
    ```

## Deployment

- Social graph
    ```bash
    kubectl -n yanqizhang apply -f socialgraph/k8s/deployment.yaml
    ```

- Post
    ```bash
    kubectl -n yanqizhang apply -f post/k8s/deployment.yaml
    ```

- User
    ```bash
    kubectl -n yanqizhang apply -f user/k8s/deployment.yaml
    ```

- Timeline update
    ```bash
    kubectl -n yanqizhang apply -f timeline-update/k8s/deployment.yaml
    ```

- Timeline read
    ```bash
    kubectl -n yanqizhang apply -f timeline-read/k8s/deployment.yaml
    ```

- Recommend
    ```bash
    kubectl -n yanqizhang apply -f recommend/k8s/deployment.yaml
    ```

- Object detect
    ```bash
    kubectl -n yanqizhang apply -f object-detect/k8s/deployment.yaml
    ```

- Sentiment
    ```bash
    kubectl -n yanqizhang apply -f sentiment/k8s/deployment.yaml
    ```

- Translate
    ```bash
    kubectl -n yanqizhang apply -f translate/k8s/deployment.yaml
    ```

- Frontend
    ```bash
    kubectl -n yanqizhang apply -f frontend/k8s/deployment.yaml
    ```

## Service

- Frontend
    Expose port 31989
    ```bash
    kubectl -n yanqizhang apply -f frontend/k8s/service_frontend.yaml
    ```

- Social graph
    Expose port 31990
    ```bash
    kubectl -n yanqizhang apply -f socialgraph/k8s/service_social.yaml
    ```

- User
    Expose port 31991
    ```bash
    kubectl -n yanqizhang apply -f user/k8s/service_user.yaml
    ```

- Recommend
    Expose port 31992
    ```bash
    kubectl -n yanqizhang apply -f recommend/k8s/service_recmd.yaml
    ```

- Translate
    Expose port 31993
    ```bash
    kubectl -n yanqizhang apply -f translate/k8s/service_transl.yaml
    ```



