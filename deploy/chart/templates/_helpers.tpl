{{/*
─── VortexUiPro Helm Template Helpers ────────────────────────────────
Functions for consistent naming and labeling across all templates.
────────────────────────────────────────────────────────────────────────
*/}}

{{- define "vortexuipro.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "vortexuipro.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "vortexuipro.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "vortexuipro.labels" -}}
helm.sh/chart: {{ include "vortexuipro.chart" . }}
{{ include "vortexuipro.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: vortexuipro
{{- end -}}

{{- define "vortexuipro.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vortexuipro.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "vortexuipro.panelSelectorLabels" -}}
app.kubernetes.io/name: {{ include "vortexuipro.name" . }}-panel
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "vortexuipro.webSelectorLabels" -}}
app.kubernetes.io/name: {{ include "vortexuipro.name" . }}-web
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "vortexuipro.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "vortexuipro.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}
