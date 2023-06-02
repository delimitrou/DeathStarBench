{{- define "hotelreservation.templates.baseServiceMemcached" }}
{{- $count := add .Values.global.memcached.HACount 1 | int }}
{{- range $key, $item := untilStep 1 $count 1 }}
{{- $rangeItem := $item -}}
{{- with $ }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Values.name }}-{{ $rangeItem }}-{{ include "hotel-reservation.fullname" . }}
spec:
  type: {{ .Values.serviceType | default .Values.global.serviceType }}
  ports:
  {{- range .Values.ports }}
  - name: "{{ .port }}"
    port: {{ .port }}
    {{- if .protocol}}
    protocol: {{ .protocol }}
    {{- end }}
    targetPort: {{ .targetPort }}
  {{- end }}
  selector:
    {{- include "hotel-reservation.selectorLabels" . | nindent 4 }}
    service: {{ .Values.name }}-{{ $rangeItem }}-{{ include "hotel-reservation.fullname" . }}
{{- end }}
{{- end }}
{{- end }}
