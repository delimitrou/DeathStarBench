{{- define "hotelreservation.templates.baseConfigMap" }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Values.name }}
  labels:
    hotelreservation/service: {{ .Values.name }}
data:
 {{- range $configMap := .Values.configMaps }}
  {{- $filePath := printf "configs/%s" $configMap.value }}
  {{ $configMap.name -}}: |
{{- tpl ($.Files.Get $filePath) $ | indent 4 -}}
  {{- end }}

{{- end }}
