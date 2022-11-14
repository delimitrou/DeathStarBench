apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Values.redis.replica.name }}
  labels:
    name: {{ .Values.redis.replica.name }}
spec:
  replicas: {{ .Values.redis.replica.count }}
  selector:
    matchLabels:
      name: {{ .Values.redis.replica.name }}
  template:
    metadata:
      labels:
        name: {{ .Values.redis.replica.name }}
    spec:
      subdomain: replica
      {{- if .Values.redis.replica.useTopologySpreadConstraint}}
      topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: node
        whenUnsatisfiable: DoNotSchedule
        labelSelector:
          matchLabels:
            name: {{ .Values.redis.replica.name }}
      {{- end}}
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
        - "--slaveof"
        - {{ .Values.redis.primary.name }}
        - "6379"
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
          name: {{ .Values.redis.replica.name }}-config
      {{- end }}