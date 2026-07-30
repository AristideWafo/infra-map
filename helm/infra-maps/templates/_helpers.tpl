{{- define "infra-maps.name" -}}
{{- .Chart.Name -}}
{{- end -}}

{{- define "infra-maps.fullname" -}}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "infra-maps.labels" -}}
app.kubernetes.io/name: {{ include "infra-maps.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "infra-maps.selectorLabels" -}}
app.kubernetes.io/name: {{ include "infra-maps.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "infra-maps.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "infra-maps.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "infra-maps.imageTag" -}}
{{- default .Chart.AppVersion .Values.image.tag -}}
{{- end -}}
