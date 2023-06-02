# Deploy benchnmark

## State stores
- Video store
    Deploy redis as video store
    ```bash
    helm install redis-video-store bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redisstore --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f video-frontend/config/video_store_redis.yaml
    ```

- Date store
    Deploy redis as date store
    ```bash
    helm install redis-date bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redisdate --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f dates/config/date_store_redis.yaml
    ```

- Rating store
    Deploy redis as rating store
    ```bash
    helm install redis-rating bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi  --set replica.persistence.size=10Gi --set global.redis.password=redisrating --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f user-rating/config/rating_store_redis.yaml
    ```

- Info store
    Deploy redis as post store
    ```bash
    helm install redis-info bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redisinfo --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f video-info/config/info_store_redis.yaml
    ```

## Pubsub

- Video pubsub
    Deploy redis as pubsub
    ```bash
    helm install redis-video-pubsub bitnami/redis -n yanqizhang -f ../helm/redis_rdb_bitnami_val.yml --set global.storageClass=local-storage --set master.persistence.size=10Gi --set replica.persistence.size=10Gi --set global.redis.password=redispubsub --wait
    ```

    Deploy dapr component
    ```bash
    kubectl -n yanqizhang apply -f video-scale/config/video_pubsub_redis.yaml
    ```

## Deployment

- Video-info
    ```bash
    kubectl -n yanqizhang apply -f video-info/k8s/deployment.yaml
    ```

- User-rating
    ```bash
    kubectl -n yanqizhang apply -f user-rating/k8s/deployment.yaml
    ```

- Trending
    ```bash
    kubectl -n yanqizhang apply -f trending/k8s/deployment.yaml
    ```

- Dates
    ```bash
    kubectl -n yanqizhang apply -f dates/k8s/deployment.yaml
    ```

- Video-scale
    ```bash
    kubectl -n yanqizhang apply -f video-scale/k8s/deployment.yaml
    ```

- Video-thumbnail
    ```bash
    kubectl -n yanqizhang apply -f video-thumbnail/k8s/deployment.yaml
    ```

- Video-frontend
    ```bash
    kubectl -n yanqizhang apply -f video-frontend/k8s/deployment.yaml
    ```

## Service

- Video-frontend
    Expose port 31985
    ```bash
    kubectl -n yanqizhang apply -f video-frontend/k8s/service_frontend.yaml
    ```

- Trending
    Expose port 31986
    ```bash
    kubectl -n yanqizhang apply -f trending/k8s/service_trending.yaml
    ```



