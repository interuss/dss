{{- define "aws-lb-default-annotations" -}}
service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing
service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: ip
service.beta.kubernetes.io/aws-load-balancer-name: {{.name}}
service.beta.kubernetes.io/aws-load-balancer-eip-allocations: {{.ip}}
service.beta.kubernetes.io/aws-load-balancer-subnets: {{.subnet}}
service.beta.kubernetes.io/aws-load-balancer-type: external
{{- end -}}

{{- define "aws-lb-crdb-annotations" -}}
{{- include "aws-lb-default-annotations" . }}
{{- end -}}

{{- define "aws-lb-spec" -}}
loadBalancerClass: service.k8s.aws/nlb
{{- end -}}

{{- define "aws-ingress-dss-gateway-annotations" -}}
{{- include "aws-lb-default-annotations" . }}
service.beta.kubernetes.io/aws-load-balancer-ssl-cert: {{.certName}}
service.beta.kubernetes.io/aws-load-balancer-ssl-ports: "443"
{{- end -}}

{{- define "aws-ingress-spec" -}}
loadBalancerClass: service.k8s.aws/nlb
{{- end -}}