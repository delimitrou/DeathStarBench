{{- define "mongodb-sharded.connection" }}
  {{ .Values.global.mongodb.sharding.svc.user }}:{{ .Values.global.mongodb.sharding.svc.password }}@{{ .Values.global.mongodb.sharding.svc.name }}
{{- end }}

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
      "addr": "social-graph-redis",
      "port": 6379,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
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
      "addr": "home-timeline-redis",
      "port": 6379,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "compose-post-service": {
      "addr": "compose-post-service",
      "port": 9090,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
    },
    "compose-post-redis": {
      "addr": "compose-post-redis",
      "port": 6379,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
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
      "addr": "user-timeline-redis",
      "port": 6379,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
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
      "addr": "post-storage-memcached",
      "port": 11211,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
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
      "addr": "media-memcached",
      "port": 11211,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
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
      "addr": "url-shorten-memcached",
      "port": 11211,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
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
      "addr": "user-memcached",
      "port": 11211,
      "connections": 512,
      "timeout_ms": 10000,
      "keepalive_ms": 10000
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
    }
  }
  {{- end }}