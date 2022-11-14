{{- define "hotelreservation.templates.basePersistentVolumeClaim" }}
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ .Values.name }}-pvc
spec:
  {{- if .Values.global.mongodb.persistentVolume.pvprovisioner.enabled }}
  {{- if .Values.global.mongodb.persistentVolume.pvprovisioner.storageClassName }}
  storageClassName: {{ .Values.global.mongodb.persistentVolume.pvprovisioner.storageClassName }}
  {{- end }}
  {{- end }}
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: {{ .Values.global.mongodb.persistentVolume.size }}
{{- end }}
