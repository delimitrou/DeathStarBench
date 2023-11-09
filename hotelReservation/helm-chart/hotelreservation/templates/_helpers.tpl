{{/*
Expand the name of the chart.
*/}}
{{- define "hotel-reservation.name" -}}
{{- default .Chart.Name .Values.global.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "hotel-reservation.fullname" -}}
{{- if .Values.global.fullnameOverride }}
{{- .Values.global.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Values.global.mainChart .Values.global.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "hotel-reservation.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "hotel-reservation.labels" -}}
helm.sh/chart: {{ include "hotel-reservation.chart" . }}
{{ include "hotel-reservation.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "hotel-reservation.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hotel-reservation.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Generate list of memcached profile service names
Usage:
  include "hotel-reservation.generateMemcAddr" (list <mapToCheck> number name port)
  e.g.
    "ProfileMemcAddress": include "hotel-reservation.generateMemcAddr" (list . 2 "memcached-profile" 11211)
*/}}
{{- define "hotel-reservation.generateMemcAddr" -}}
  {{- $mapToCheck := index . 0 }}
  {{- $count := add (index . 1) 1 | int }}
  {{- $name := index . 2 }}
  {{- $port := index . 3 | int }}
  {{- $fullname := include "hotel-reservation.fullname" $mapToCheck }}
  {{- $appendix := printf "%s.%s.svc.%s:%d" $fullname $mapToCheck.Release.Namespace $mapToCheck.Values.global.serviceDnsDomain $port }}
  {{- $addrlist := list }}
  {{- range $key, $item := untilStep 1 $count 1 }}
    {{- $addr := printf "%s-%d-%s" $name $item $appendix }}
    {{- $addrlist = append $addrlist $addr }}
  {{- end }}
  {{- join "," $addrlist | toJson }}
{{- end }}

{{/*
Backend labels
*/}}
{{- define "hotel-reservation.backendLabels" -}}
backend: {{ .Chart.Name | regexFind "[^-]*" }}
{{- end }}
