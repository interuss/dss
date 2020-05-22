module github.com/interuss/dss

// This forked version of openapi2proto has limited support for Open API v3.
replace github.com/NYTimes/openapi2proto => github.com/davidsansome/openapi2proto v0.2.3-0.20190826092301-b98d13b38dab

go 1.14

require (
	cloud.google.com/go v0.57.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/cockroachdb/cockroach-go v0.0.0-20200504194139-73ffeee90b62
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dpjacques/clockwork v0.1.0
	github.com/gogo/protobuf v1.3.1
	github.com/gogo/status v1.1.0
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.5
	github.com/lib/pq v1.5.2
	github.com/stretchr/testify v1.5.1
	go.uber.org/multierr v1.5.0
	go.uber.org/zap v1.15.0
	google.golang.org/genproto v0.0.0-20200519141106-08726f379972
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.23.0
	gopkg.in/square/go-jose.v2 v2.5.1
)
