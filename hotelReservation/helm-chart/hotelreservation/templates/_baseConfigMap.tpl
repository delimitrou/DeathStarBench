{{- define "hotelreservation.templates.baseConfigMap" }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}
  labels:
    hotelreservation/service: {{ .Values.name }}--{{ include "hotel-reservation.fullname" . }}
data:
 {{- range $configMap := .Values.configMaps }}
  {{- $filePath := printf "configs/%s" $configMap.value }}
  {{ $configMap.name -}}: |
{{- tpl ($.Files.Get $filePath) $ | indent 4 -}}
  {{- end }}

{{- end }}
