{{- define "mongodb-sharded.connection" }}
  {{ .Values.global.mongodb.sharding.svc.user }}:{{ .Values.global.mongodb.sharding.svc.password }}@{{ .Values.global.mongodb.sharding.svc.name }}
{{- end }}

{{- define "memcached-cluster.connection" }}
  {{ .Release.Name }}-mcrouter
{{- end }}

{{- define "redis-cluster.connection" }}
  {{ .Release.Name }}-redis-cluster
{{- end}}

{{- define "socialnetwork.templates.other.service-config.json"  }}
{
    "secret": "secret",
    "social-graph-service": {
      "addr": "social-graph-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "social-graph-mongodb": {
      "addr": {{ ternary (include "mongodb-sharded.connection" . | trim) "social-graph-mongodb" .Values.global.mongodb.sharding.enabled | quote}},
      "port": {{ ternary .Values.global.mongodb.sharding.svc.port 27017 .Values.global.mongodb.sharding.enabled}},
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "social-graph-redis": {
      "addr": {{ ternary (include "redis-cluster.connection" . | trim) "social-graph-redis" .Values.global.redis.cluster.enabled | quote}},
      "port": 6379,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000,
      "use_cluster": {{ ternary 1 0 .Values.global.redis.cluster.enabled}},
      "use_replica": {{ ternary 1 0 .Values.global.redis.replication.enabled}}
    },
    "write-home-timeline-service": {
      "addr": "write-home-timeline-service",
      "port": 9090,
      "workers": 32,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "write-home-timeline-rabbitmq": {
      "addr": "write-home-timeline-rabbitmq",
      "port": 5672,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "home-timeline-redis": {
      "addr": {{ ternary (include "redis-cluster.connection" . | trim) "home-timeline-redis" .Values.global.redis.cluster.enabled | quote}},
      "port": 6379,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000,
      "use_cluster": {{ ternary 1 0 .Values.global.redis.cluster.enabled}},
      "use_replica": {{ ternary 1 0 .Values.global.redis.replication.enabled}}
    },
    "compose-post-service": {
      "addr": "compose-post-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "compose-post-redis": {
      "addr": {{ ternary (include "redis-cluster.connection" . | trim) "compose-post-redis" .Values.global.redis.cluster.enabled | quote}},
      "port": 6379,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000,
      "use_cluster": {{ ternary 1 0 .Values.global.redis.cluster.enabled}},
      "use_replica": {{ ternary 1 0 .Values.global.redis.replication.enabled}}
    },
    "user-timeline-service": {
      "addr": "user-timeline-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "user-timeline-mongodb": {
      "addr": {{ ternary (include "mongodb-sharded.connection" . | trim) "user-timeline-mongodb" .Values.global.mongodb.sharding.enabled | quote}},
      "port": {{ ternary .Values.global.mongodb.sharding.svc.port 27017 .Values.global.mongodb.sharding.enabled}},
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "user-timeline-redis": {
      "addr": {{ ternary (include "redis-cluster.connection" . | trim) "user-timeline-redis" .Values.global.redis.cluster.enabled | quote}},
      "port": 6379,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000,
      "use_cluster": {{ ternary 1 0 .Values.global.redis.cluster.enabled}},
      "use_replica": {{ ternary 1 0 .Values.global.redis.replication.enabled}}
    },
    "post-storage-service": {
      "addr": "post-storage-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "post-storage-mongodb": {
      "addr": {{ ternary (include "mongodb-sharded.connection" . | trim) "post-storage-mongodb" .Values.global.mongodb.sharding.enabled | quote}},
      "port": {{ ternary .Values.global.mongodb.sharding.svc.port 27017 .Values.global.mongodb.sharding.enabled}},
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "post-storage-memcached": {
      "addr": {{ ternary (include "memcached-cluster.connection" . | trim) "post-storage-memcached" .Values.global.memcached.cluster.enabled | quote}},
      "port": {{ ternary .Values.global.memcached.cluster.port 11211 .Values.global.memcached.cluster.enabled}},
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000,
      "binary_protocol": {{ ternary 0 1 .Values.global.memcached.cluster.enabled}}
    },
    "unique-id-service": {
      "addr": "unique-id-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000,
      "netif": "eth0"
    },
    "media-service": {
      "addr": "media-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "media-mongodb": {
      "addr": {{ ternary (include "mongodb-sharded.connection" . | trim) "media-mongodb" .Values.global.mongodb.sharding.enabled | quote}},
      "port": {{ ternary .Values.global.mongodb.sharding.svc.port 27017 .Values.global.mongodb.sharding.enabled}},
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "media-memcached": {
      "addr": {{ ternary (include "memcached-cluster.connection" . | trim) "media-memcached" .Values.global.memcached.cluster.enabled | quote}},
      "port": {{ ternary .Values.global.memcached.cluster.port 11211 .Values.global.memcached.cluster.enabled}},
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000,
      "binary_protocol": {{ ternary 0 1 .Values.global.memcached.cluster.enabled}}
    },
    "media-frontend": {
      "addr": "media-frontend",
      "port": 8081,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "text-service": {
      "addr": "text-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "user-mention-service": {
      "addr": "user-mention-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "url-shorten-service": {
      "addr": "url-shorten-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "url-shorten-memcached": {
      "addr": {{ ternary (include "memcached-cluster.connection" . | trim) "url-shorten-memcached" .Values.global.memcached.cluster.enabled | quote}},
      "port": {{ ternary .Values.global.memcached.cluster.port 11211 .Values.global.memcached.cluster.enabled}},
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000,
      "binary_protocol": {{ ternary 0 1 .Values.global.memcached.cluster.enabled}}
    },
    "url-shorten-mongodb": {
      "addr": {{ ternary (include "mongodb-sharded.connection" . | trim) "url-shorten-mongodb" .Values.global.mongodb.sharding.enabled | quote}},
      "port": {{ ternary .Values.global.mongodb.sharding.svc.port 27017 .Values.global.mongodb.sharding.enabled}},
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "user-service": {
      "addr": "user-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000,
      "netif": "eth0"
    },
    "user-memcached": {
      "addr": {{ ternary (include "memcached-cluster.connection" . | trim) "user-memcached" .Values.global.memcached.cluster.enabled | quote}},
      "port": {{ ternary .Values.global.memcached.cluster.port 11211 .Values.global.memcached.cluster.enabled}},
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000,
      "binary_protocol": {{ ternary 0 1 .Values.global.memcached.cluster.enabled}}
    },
    "user-mongodb": {
      "addr": {{ ternary (include "mongodb-sharded.connection" . | trim) "user-mongodb" .Values.global.mongodb.sharding.enabled | quote}},
      "port": {{ ternary .Values.global.mongodb.sharding.svc.port 27017 .Values.global.mongodb.sharding.enabled}},
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "home-timeline-service": {
      "addr": "home-timeline-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "ssl": {
      "enabled": false,
      "caPath": "/keys/CA.pem",
      "ciphers": "ALL:!ADH:!LOW:!EXP:!MD5:@STRENGTH",
      "serverKeyPath": "/keys/server.key",
      "serverCertPath": "/keys/server.crt"
    },
    "redis-primary": {
      "keepalive_ms": 10000,
      "addr": {{ .Values.global.redis.replication.primary | quote }},
      "timeout_ms": 10000,
      "port": 6379,
      "connections": 512
    },
    "redis-replica": {
      "keepalive_ms": 10000,
      "addr": {{ .Values.global.redis.replication.replica | quote }},
      "timeout_ms": 10000,
      "port": 6379,
      "connections": 512
    }
  }
  {{- end }}