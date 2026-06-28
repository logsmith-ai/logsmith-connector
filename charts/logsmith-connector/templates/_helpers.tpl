{{/* Expand the name of the chart. */}}
{{- define "logsmith-connector.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Fully qualified app name. */}}
{{- define "logsmith-connector.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name (include "logsmith-connector.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/* Common labels. */}}
{{- define "logsmith-connector.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "logsmith-connector.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/* Selector labels. */}}
{{- define "logsmith-connector.selectorLabels" -}}
app.kubernetes.io/name: {{ include "logsmith-connector.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/* Name of the Secret holding the token. */}}
{{- define "logsmith-connector.secretName" -}}
{{- if .Values.existingSecret -}}
{{- .Values.existingSecret -}}
{{- else -}}
{{- include "logsmith-connector.fullname" . -}}
{{- end -}}
{{- end -}}

{{/* Key within the Secret holding the token. */}}
{{- define "logsmith-connector.secretKey" -}}
{{- if .Values.existingSecret -}}
{{- .Values.existingSecretKey | default "token" -}}
{{- else -}}
token
{{- end -}}
{{- end -}}
