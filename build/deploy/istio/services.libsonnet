{
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: cockroachdb-gateway
spec:
  selector:
    istio: ingressgateway # use Istio default gateway implementation
  servers:
  - port:
      number: 26257
      name: cockroach
      protocol: TCP
    hosts:
    - "*.db.interussplatform.dev"
    tls:
      mode: MUTUAL
}