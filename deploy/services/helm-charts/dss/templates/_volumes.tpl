{{- define "ca-certs:volume" -}}
- name: ca-certs
  secret:
    defaultMode: 256
    secretName: {{ if .cockroachdb }}cockroachdb.ca.crt{{ else }}yugabyte-tls-client-cert{{ end }}
{{- end -}}
{{- define "ca-certs:volumeMount" -}}
{{ if .cockroachdb }}
- mountPath: /cockroach/cockroach-certs/ca.crt
  name: ca-certs
  subPath: ca.crt
{{ else }}
- mountPath: /opt/yugabyte-certs/ca.crt
  name: ca-certs
  subPath: root.crt
{{- end -}}
{{- end -}}
{{- define "client-certs:volume" -}}
- name: client-certs
  secret:
    defaultMode: 256
    secretName: {{ if .cockroachdb }}cockroachdb.client.root{{ else }}yugabyte-tls-client-cert{{ end }}
{{- end -}}
{{- define "client-certs:volumeMount" -}}
{{ if .cockroachdb }}
- mountPath: /cockroach/cockroach-certs/client.root.crt
  name: client-certs
  subPath: client.root.crt
- mountPath: /cockroach/cockroach-certs/client.root.key
  name: client-certs
  subPath: client.root.key
{{ else }}
- mountPath: /opt/yugabyte-certs/client.yugabyte.crt
  name: client-certs
  subPath: yugabytedb.crt
- mountPath: /opt/yugabyte-certs/client.yugabyte.key
  name: client-certs
  subPath: yugabytedb.key
{{- end -}}
{{- end -}}


{{- define "public-certs:volume" -}}
- name: public-certs
  secret:
    defaultMode: 256
    secretName: dss.public.certs
{{- end -}}
{{- define "public-certs:volumeMount" -}}
- mountPath: /public-certs
  name: public-certs
{{- end -}}
