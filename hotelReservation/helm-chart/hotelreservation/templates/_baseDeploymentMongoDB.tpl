{{- define "hotelreservation.templates.baseDeploymentMongoDB" }}
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    {{- include "hotel-reservation.labels" . | nindent 4 }}
    {{- include "hotel-reservation.backendLabels" . | nindent 4 }}
    service: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}
  name: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}
spec:
  replicas: {{ .Values.replicas | default .Values.global.replicas }}
  selector:
    matchLabels:
      {{- include "hotel-reservation.selectorLabels" . | nindent 6 }}
      {{- include "hotel-reservation.backendLabels" . | nindent 6 }}
      service: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}
      app: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}
  template:
    metadata:
      labels:
        {{- include "hotel-reservation.labels" . | nindent 8 }}
        {{- include "hotel-reservation.backendLabels" . | nindent 8 }}
        service: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}
        app: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}
      {{- if hasKey $.Values "annotations" }}
      annotations:
        {{ tpl $.Values.annotations . | nindent 8 | trim }}
      {{- else if hasKey $.Values.global "annotations" }}
      annotations:
        {{ tpl $.Values.global.annotations . | nindent 8 | trim }}
      {{- end }}
    spec:
      containers:
      {{- with .Values.container }}
      - name: "{{ .name }}"
        image: {{ .dockerRegistry | default $.Values.global.dockerRegistry }}/{{ .image }}:{{ .imageVersion | default $.Values.global.defaultImageVersion }}
        imagePullPolicy: {{ .imagePullPolicy | default $.Values.global.imagePullPolicy }}
        ports:
        {{- range $cport := .ports }}
        - containerPort: {{ $cport.containerPort -}}
        {{ end }}
        {{- if .command}}
        command:
        - {{ .command }}
        {{- end -}}
        {{- if .args}}
        args:
        {{- range $arg := .args}}
        - {{ $arg }}
        {{- end -}}
        {{- end }}
        {{- if .resources }}
        resources:
          {{ tpl .resources $ | nindent 10 | trim }}
        {{- else if hasKey $.Values.global "resources" }}
        resources:
          {{ tpl $.Values.global.resources $ | nindent 10 | trim }}
        {{- end }}
        volumeMounts:
        - mountPath: /data/db
          name: {{ $.Values.name }}-{{ include "hotel-reservation.fullname" $ }}-path
      {{- end }}
      volumes:
      - name: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}-path
	{{- if $.Values.global.mongodb.persistentVolume.enabled }}
        persistentVolumeClaim:
          claimName: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}-pvc
        {{- else }}
        emptyDir: {}
        {{- end }}
      {{- if hasKey .Values "topologySpreadConstraints" }}
      topologySpreadConstraints:
        {{ tpl .Values.topologySpreadConstraints . | nindent 6 | trim }}
      {{- else if hasKey $.Values.global.mongodb "topologySpreadConstraints" }}
      topologySpreadConstraints:
        {{ tpl $.Values.global.mongodb.topologySpreadConstraints . | nindent 6 | trim }}
      {{- end }}
      hostname: {{ .Values.name }}-{{ include "hotel-reservation.fullname" . }}
      restartPolicy: {{ .Values.restartPolicy | default .Values.global.restartPolicy}}
      {{- if .Values.affinity }}
      affinity: {{- toYaml .Values.affinity | nindent 8 }}
      {{- else if hasKey $.Values.global "affinity" }}
      affinity: {{- toYaml .Values.global.affinity | nindent 8 }}
      {{- end }}
      {{- if .Values.tolerations }}
      tolerations: {{- toYaml .Values.tolerations | nindent 8 }}
      {{- else if hasKey $.Values.global "tolerations" }}
      tolerations: {{- toYaml .Values.global.tolerations | nindent 8 }}
      {{- end }}
      {{- if .Values.nodeSelector }}
      nodeSelector: {{- toYaml .Values.nodeSelector | nindent 8 }}
      {{- else if hasKey $.Values.global "nodeSelector" }}
      nodeSelector: {{- toYaml .Values.global.nodeSelector | nindent 8 }}
      {{- end }}
{{- end}}
