{{- define "mediamicroservices.templates.nginx.jaeger-config.json"  }}

{
  "service_name": "nginx-web-server",
  "disabled": {{ .Values.global.jaeger.disabled }},
  "reporter": {
    "logSpans": {{ .Values.global.jaeger.logSpans }},
    "localAgentHostPort": "{{ .Values.global.jaeger.localAgentHostPort }}",
    "queueSize": {{ int .Values.global.jaeger.queueSize }},
    "bufferFlushInterval": {{ int .Values.global.jaeger.bufferFlushInterval }}
  },
  "sampler": {
    "type": "{{ .Values.global.jaeger.samplerType }}",
    "param": {{ .Values.global.jaeger.samplerParam }}
  }
}

{{- end }}