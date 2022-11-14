apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Values.redis.primary.name }}
  labels:
    name: {{ .Values.redis.primary.name }}
spec:
  replicas: 1  # only single primary is supported
  selector:
    matchLabels:
      name: {{ .Values.redis.primary.name }}
  template:
    metadata:
      labels:
        name: {{ .Values.redis.primary.name }}
    spec:
      subdomain: primary
      containers:
      - name: redis
        image: {{ .Values.redis.registry }}/{{ .Values.redis.image }}:{{ .Values.redis.tag }}
        command:
        - "redis-server"
        {{- if .Values.redis.useConfigmap }}
        args:
        - /usr/local/etc/redis/redis.conf
        {{- else }}
        args:
        - "--protected-mode"
        - "no"
        {{- end }}
        {{- if .Values.redis.useConfigmap }}
        volumeMounts:
        - name: redis-config
          mountPath: /usr/local/etc/redis/
        {{- end }}
        ports:
        - containerPort: 6379
      {{- if .Values.redis.useConfigmap }}
      volumes:
      - name: redis-config
        configMap:
          name: {{ .Values.redis.primary.name }}-config
      {{- end }}