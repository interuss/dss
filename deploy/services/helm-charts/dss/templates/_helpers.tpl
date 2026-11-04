{{- define "datastoreImage" -}}
{{- if $.Values.cockroachdb.enabled -}}
{{ (printf "%s:%s" $.Values.cockroachdb.image.repository $.Values.cockroachdb.image.tag) }}
{{- else -}}
{{ (printf "%s:%s" $.Values.yugabyte.Image.repository $.Values.yugabyte.Image.tag) }}
{{- end -}}
{{- end -}}

{{- define "datastorePort" -}}
{{- if $.Values.cockroachdb.enabled -}}
26257
{{- else -}}
5433
{{- end -}}
{{- end -}}

{{- define "datastoreUser" -}}
{{- if $.Values.cockroachdb.enabled -}}
root
{{- else -}}
yugabyte
{{- end -}}
{{- end -}}


{{- define "datastoreHost" -}}
{{- if $.Values.cockroachdb.enabled -}}
{{- printf "%s-public.%s" $.Values.cockroachdb.fullnameOverride $.Release.Namespace -}}
{{- else -}}
{{- printf "yb-tservers.%s" $.Release.Namespace -}}
{{- end -}}
{{- end -}}

{{- define "init-container-wait-for-http" -}}
- name: wait-for-{{.serviceName}}
  image: alpine:3.17.3
  command: [ 'sh', '-c', "until wget -nv {{.url}}; do echo waiting for {{.serviceName}}; sleep 2; done" ]
{{- end -}}

{{- define "init-container-wait-for-schema" -}}
{{/*For some reason, calling the template datastoreImage fails here.*/}}
- name: wait-for-schema-{{.schemaName}}
  image: {{.datastoreImage}}
  volumeMounts:
    {{- include "ca-certs:volumeMount" $ | nindent 4 }}
    {{- include "client-certs:volumeMount" $ | nindent 4 }}
  command:
    - sh
    - -c
{{ if .cockroachdbEnabled }}
    - "/cockroach/cockroach sql --certs-dir /cockroach/cockroach-certs/ --host {{.datastoreHost}} --port \"{{.datastorePort}}\" --format raw -e \"SELECT * FROM crdb_internal.databases where name = '{{.schemaName}}';\" | grep {{.schemaName}}"
{{ else }}
    - "ysqlsh --host {{.datastoreHost}} --port \"{{.datastorePort}}\" \"sslmode=require sslcert=/opt/yugabyte-certs/client.yugabyte.crt sslkey=/opt/yugabyte-certs/client.yugabyte.key sslrootcert=/opt/yugabyte-certs/ca.crt\" -c \"SELECT datname FROM pg_database where datname = '{{.schemaName}}';\" | grep {{.schemaName}}"
{{ end }}
{{- end -}}

{{/* Common core-service CLI params of the SQL and Raft deployments. */}}
{{- define "core-service-args" -}}
{{- $conf := .Values.dss.conf -}}
{{- $params := list
  (dict "--accepted_jwt_audiences" $conf.hostname)
  (dict "--addr" ":8080")
  (dict "--public_endpoint" (ternary $conf.publicEndpoint nil (not (empty $conf.publicEndpoint))))
  (dict "--locality" $conf.locality)
  (dict "--enable_scd" ($conf.enableScd | default false | toString))
  (dict "--enable_scd_global_lock" ($conf.enableScdGlobalLock | default false | toString))
  (dict "--enable_time_based_notification_index" (ternary "true" nil ($conf.enableTimeBasedNotificationIndex | default false)))
  (dict "--enable_metrics" (ternary "true" nil ($conf.enableDssMetrics | default false)))
  (dict "--gcp_prof_service_name" ($conf.gcpProfServiceName | default ""))
  (dict "--jwks_endpoint" (ternary $conf.jwksEndpoint nil (not (empty $conf.jwksEndpoint))))
  (dict "--jwks_key_ids" (ternary ($conf.jwksKeyIds | join ",") nil (not (empty $conf.jwksKeyIds))))
  (dict "--jwks_refresh_interval" (ternary $conf.jwksRefreshInterval nil (not (empty $conf.jwksRefreshInterval))))
  (dict "--jwks_key_ttl" (ternary $conf.jwksKeyTtl nil (not (empty $conf.jwksKeyTtl))))
  (dict "--public_key_files" (ternary ($conf.pubKeys | join ",") nil (not (empty $conf.pubKeys))))
-}}
{{- $lines := list -}}
{{- range $params -}}
{{- range $flag, $value := . -}}
{{- if not (kindIs "invalid" $value) -}}
{{- $lines = append $lines (printf "%s=%s" $flag (toString $value)) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- if not $lines -}}
{{- fail "core-service-args: at least one parameter must have a non-null value" -}}
{{- end -}}
{{- join "\n" $lines -}}
{{- end -}}
