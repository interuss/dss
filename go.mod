module github.com/interuss/dss

// This forked version of openapi2proto has limited support for Open API v3.
replace github.com/NYTimes/openapi2proto => github.com/davidsansome/openapi2proto v0.2.3-0.20190826092301-b98d13b38dab

go 1.14

require (
	cloud.google.com/go v0.64.0
	github.com/cockroachdb/cockroach-go v0.0.0-20200504194139-73ffeee90b62
	github.com/coreos/go-semver v0.3.0
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/interuss/stacktrace v1.0.0
	github.com/jonboulle/clockwork v0.2.2
	github.com/lib/pq v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.16.0
	google.golang.org/genproto v0.0.0-20201030142918-24207fddd1c3
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/square/go-jose.v2 v2.5.1
)
