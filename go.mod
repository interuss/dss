module github.com/interuss/dss

// This forked version of openapi2proto has limited support for Open API v3.
replace github.com/NYTimes/openapi2proto => github.com/davidsansome/openapi2proto v0.2.3-0.20190826092301-b98d13b38dab

go 1.12

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/jonboulle/clockwork v0.1.0
	github.com/lib/pq v1.2.0
	go.uber.org/multierr v1.4.0
	go.uber.org/zap v1.13.0
	google.golang.org/genproto v0.0.0-20191115221424-83cc0476cb11
	google.golang.org/grpc v1.25.1
	gopkg.in/square/go-jose.v2 v2.4.0
	gopkg.in/yaml.v2 v2.2.7
	k8s.io/apimachinery v0.0.0-20191121015412-41065c7a8c2a
)
