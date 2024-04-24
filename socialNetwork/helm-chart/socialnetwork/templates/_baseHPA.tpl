{{- define "socialnetwork.templates.baseHPA" }}

{{- if or (and .Values.hpa .Values.hpa.enabled) ($.Values.global.hpa.enabled) -}}
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name:  {{ .Values.name }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ .Values.name }}
  {{- if and .Values.hpa .Values.hpa.minReplicas}}
  minReplicas: {{ .Values.hpa.minReplicas }}
  {{- else}}
  minReplicas: {{ default 1 .Values.global.hpa.minReplicas }}
  {{- end}}
  {{- if and .Values.hpa .Values.hpa.maxReplicas}}
  maxReplicas: {{ .Values.hpa.maxReplicas }}
  {{- else}}
  maxReplicas: {{ default 1 .Values.global.hpa.maxReplicas }}
  {{- end}}
  metrics:
  {{- if or $.Values.global.hpa.targetMemoryUtilizationPercentage (and .Values.hpa .Values.hpa.targetMemoryUtilizationPercentage) }}  
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          {{- if and .Values.hpa .Values.hpa.targetMemoryUtilizationPercentage }}
          averageUtilization: {{ .Values.hpa.targetMemoryUtilizationPercentage }}
          {{- else }}
          averageUtilization: {{ $.Values.global.hpa.targetMemoryUtilizationPercentage }}
          {{- end}}
  {{- end }}
  {{- if or $.Values.global.hpa.targetCPUUtilizationPercentage (and .Values.hpa .Values.hpa.targetCPUUtilizationPercentage) }}  
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          {{- if and .Values.hpa .Values.hpa.targetCPUUtilizationPercentage }}
          averageUtilization: {{ .Values.hpa.targetCPUUtilizationPercentage }}
          {{- else }}
          averageUtilization: {{ $.Values.global.hpa.targetCPUUtilizationPercentage }}
          {{- end}}
  {{- end }}
{{- end }}
{{- end }}
