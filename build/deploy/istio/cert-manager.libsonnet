apiVersion: authentication.istio.io/v1alpha1
kind: Policy
metadata:
  name: default
  namespace: certificate
spec:
  peers:
    - mtls:
        mode: PERMISSIVE
---
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: disable-mtls
  namespace: certificate
spec:
  host: "*.certificate.svc.cluster.local"
  trafficPolicy:
    tls:
      mode: DISABLE
