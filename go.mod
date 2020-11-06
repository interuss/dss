module github.com/interuss/dss

// This forked version of openapi2proto has limited support for Open API v3.
replace github.com/NYTimes/openapi2proto => github.com/davidsansome/openapi2proto v0.2.3-0.20190826092301-b98d13b38dab

go 1.14

require (
	cloud.google.com/go v0.57.0
	github.com/cockroachdb/cockroach-go v0.0.0-20200504194139-73ffeee90b62
	github.com/cockroachdb/cockroach-go/v2 v2.0.8
	github.com/coreos/go-semver v0.3.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dpjacques/clockwork v0.1.0
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.3
	github.com/interuss/stacktrace v0.0.0-20200827180054-b2e58cf48818
	github.com/jackc/pgconn v1.5.0
	github.com/jackc/pgtype v1.3.0
	github.com/jackc/pgx/v4 v4.6.0
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177 // indirect
	github.com/robfig/cron/v3 v3.0.1
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897 // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20200519141106-08726f379972
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.23.0
	gopkg.in/square/go-jose.v2 v2.5.1
)
