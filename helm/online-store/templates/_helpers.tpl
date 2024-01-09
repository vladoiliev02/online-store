{{/* Generate the labels for the chart */}}
{{- define "online-store.labels" -}}
app: online-store
environment: {{ .Values.environment | default "development" }}
{{- end }}