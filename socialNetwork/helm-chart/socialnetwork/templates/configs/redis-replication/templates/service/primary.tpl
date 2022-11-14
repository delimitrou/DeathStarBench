apiVersion: v1
kind: Service
metadata:
  name: {{ .Values.redis.primary.name }}
spec:
  ports:
  - protocol: TCP
    port: 6379
    targetPort: 6379
    name: redis
  selector:
    name: {{ .Values.redis.primary.name }}