{{- define "ca-certs:volume" -}}
- name: ca-certs
  secret:
    defaultMode: 256
    secretName: cockroachdb.ca.crt
{{- end -}}
{{- define "ca-certs:volumeMount" -}}
- mountPath: /cockroach/cockroach-certs/ca.crt
  name: ca-certs
  subPath: ca.crt
{{- end -}}

{{- define "client-certs:volume" -}}
- name: client-certs
  secret:
    defaultMode: 256
    secretName: cockroachdb.client.root
{{- end -}}
{{- define "client-certs:volumeMount" -}}
- mountPath: /cockroach/cockroach-certs/client.root.crt
  name: client-certs
  subPath: client.root.crt
- mountPath: /cockroach/cockroach-certs/client.root.key
  name: client-certs
  subPath: client.root.key
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