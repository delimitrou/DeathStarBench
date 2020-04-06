for i in cast-info-memcached cast-info-mongodb cast-info-service compose-review-memcached compose-review-service jaeger movie-id-memcached movie-id-mongodb movie-id-service movie-info-memcached movie-info-mongodb movie-info-service movie-review-mongodb movie-review-redis movie-review-service nginx-web-server plot-memcached plot-mongodb plot-service rating-redis rating-service review-storage-memcached review-storage-mongodb review-storage-service text-service unique-id-service user-memcached user-mongodb user-review-mongodb user-review-redis user-review-service user-service mms-client
do
	echo "apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: ${i}
spec:
  host: ${i}.media-microsvc.svc.cluster.local
  subsets:
  - name: ${i}
    labels:
      app: ${i}
---"

done

