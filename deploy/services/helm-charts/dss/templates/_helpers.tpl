{{- define "cockroachImage" -}}
{{ (printf "%s:%s" $.Values.cockroachdb.image.repository $.Values.cockroachdb.image.tag) }}
{{- end -}}

{{- define "cockroachHost" -}}
{{- printf "%s-public.default" $.Values.cockroachdb.fullnameOverride -}}
{{- end -}}

{{- define "init-container-wait-for-http" -}}
- name: wait-for-{{.serviceName}}
  image: alpine:3.17.3
  command: [ 'sh', '-c', "until wget -nv {{.url}}; do echo waiting for {{.serviceName}}; sleep 2; done" ]
{{- end -}}

{{- define "init-container-wait-for-schema" -}}
{{/*For some reason, calling the template cockroachImage fails here.*/}}
- name: wait-for-schema-{{.schemaName}}
  image: {{.cockroachImage}}
  volumeMounts:
    {{- include "ca-certs:volumeMount" . | nindent 4 }}
    {{- include "client-certs:volumeMount" . | nindent 4 }}
  command:
    - sh
    - -c
    - "/cockroach/cockroach sql --certs-dir /cockroach/cockroach-certs/ --host {{.cockroachHost}} --port \"26257\" --format raw -e \"SELECT * FROM crdb_internal.databases where name = '{{.schemaName}}';\" | grep {{.schemaName}}"
{{- end -}}