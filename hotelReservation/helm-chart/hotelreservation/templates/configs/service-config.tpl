{{- define "hotelreservation.templates.service-config.json" }}
{
    "consulAddress": "consul.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:8500",
    "jaegerAddress": "jaeger.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:6831",
    "FrontendPort": "5000",
    "GeoPort": "8083",
    "GeoMongoAddress": "mongodb-geo.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:27018",
    "ProfilePort": "8081",
    "ProfileMongoAddress": "mongodb-profile.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:27019",
    "ProfileMemcAddress": "memcached-profile.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:11213",
    "RatePort": "8084",
    "RateMongoAddress": "mongodb-rate.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:27020",
    "RateMemcAddress": "memcached-rate.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:11212",
    "RecommendPort": "8085",
    "RecommendMongoAddress": "mongodb-recommendation.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:27021",
    "ReservePort": "8087",
    "ReserveMongoAddress": "mongodb-reservation.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:27022",
    "ReserveMemcAddress": "memcached-reserve.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:11214",
    "SearchPort": "8082",
    "UserPort": "8086",
    "UserMongoAddress": "mongodb-user.{{ .Release.Namespace }}.svc.{{ .Values.global.serviceDnsDomain }}:27023"
}
{{- end }}

