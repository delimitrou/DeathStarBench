{{- define "hotelreservation.templates.basePersistentVolumeClaim" }}
{{- if .Values.global.mongodb.persistentVolume.pvprovisioner.enabled }}
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}-pvc
spec:
  {{- if .Values.global.mongodb.persistentVolume.pvprovisioner.storageClassName }}
  storageClassName: {{ .Values.global.mongodb.persistentVolume.pvprovisioner.storageClassName }}
  {{- end }}
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: {{ .Values.global.mongodb.persistentVolume.size }}
{{- end }}
{{- end }}
