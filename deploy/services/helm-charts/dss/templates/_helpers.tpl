{{- define "databaseImage" -}}
{{- if $.Values.cockroachdb.enabled -}}
{{ (printf "%s:%s" $.Values.cockroachdb.image.repository $.Values.cockroachdb.image.tag) }}
{{- else -}}
{{ (printf "%s:%s" $.Values.yugabyte.Image.repository $.Values.yugabyte.Image.tag) }}
{{- end -}}
{{- end -}}

{{- define "databasePort" -}}
{{- if $.Values.cockroachdb.enabled -}}
26257
{{- else -}}
5433
{{- end -}}
{{- end -}}

{{- define "databaseUser" -}}
{{- if $.Values.cockroachdb.enabled -}}
root
{{- else -}}
yugabyte
{{- end -}}
{{- end -}}


{{- define "databaseHost" -}}
{{- if $.Values.cockroachdb.enabled -}}
{{- printf "%s-public.default" $.Values.cockroachdb.fullnameOverride -}}
{{- else -}}
{{- printf "yb-tservers.default" -}}
{{- end -}}
{{- end -}}

{{- define "init-container-wait-for-http" -}}
- name: wait-for-{{.serviceName}}
  image: alpine:3.17.3
  command: [ 'sh', '-c', "until wget -nv {{.url}}; do echo waiting for {{.serviceName}}; sleep 2; done" ]
{{- end -}}

{{- define "init-container-wait-for-schema" -}}
{{/*For some reason, calling the template databaseImage fails here.*/}}
- name: wait-for-schema-{{.schemaName}}
  image: {{.databaseImage}}
  volumeMounts:
    {{- include "ca-certs:volumeMount" . | nindent 4 }}
    {{- include "client-certs:volumeMount" . | nindent 4 }}
  command:
    - sh
    - -c
{{ if .cockroachdb }}
    - "/cockroach/cockroach sql --certs-dir /cockroach/cockroach-certs/ --host {{.databaseHost}} --port \"{{.databasePort}}\" --format raw -e \"SELECT * FROM crdb_internal.databases where name = '{{.schemaName}}';\" | grep {{.schemaName}}"
{{ else }}
    - "ysqlsh  --host {{.databaseHost}} --port \"{{.databasePort}}\" -c \"SELECT datname FROM pg_database where datname = '{{.schemaName}}';\" | grep {{.schemaName}}"
{{ end }}
{{- end -}}
