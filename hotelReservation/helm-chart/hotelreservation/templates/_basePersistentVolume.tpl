{{- define "hotelreservation.templates.basePersistentVolume" }}
{{- if .Values.global.mongodb.persistentVolume.enabled }}
{{- if .Values.global.mongodb.persistentVolume.hostPath.enabled }}
apiVersion: v1
kind: PersistentVolume
metadata:
  name: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}-pv
  labels:
    app-name: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}
spec:
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: {{ .Values.global.mongodb.persistentVolume.size }}
  hostPath:
    path: {{ .Values.global.mongodb.persistentVolume.hostPath.path }}/{{ .Values.name }}-pv
    type: DirectoryOrCreate
{{- end }}
{{- end }}
{{- end }}
