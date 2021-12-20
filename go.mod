module github.com/interuss/dss

// This forked version of openapi2proto has limited support for Open API v3.
replace github.com/NYTimes/openapi2proto => github.com/davidsansome/openapi2proto v0.2.3-0.20190826092301-b98d13b38dab

go 1.17

require (
	cloud.google.com/go v0.88.0
	github.com/cockroachdb/cockroach-go v0.0.0-20200504194139-73ffeee90b62
	github.com/coreos/go-semver v0.3.0
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang-migrate/migrate/v4 v4.15.1
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/interuss/stacktrace v1.0.0
	github.com/jonboulle/clockwork v0.2.2
	github.com/lib/pq v1.10.0
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.16.0
	google.golang.org/genproto v0.0.0-20211013025323-ce878158c4d4
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/square/go-jose.v2 v2.5.1
)

require (
	github.com/cockroachdb/cockroach-go/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/pprof v0.0.0-20210715191844-86eeefc3e471 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20211013171255-e13a2654a71e // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914 // indirect
	golang.org/x/sys v0.0.0-20211013075003-97ac67df715c // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/api v0.51.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

// This replacement should actually be for v3.2.1 to fix the second-level
// dependency issue with go-migrate eventually depending on the obsolete
// dgrijalva package, but Go will not allow this giving an error of:
//   go: github.com/golang-jwt/jwt@v3.2.1+incompatible used for two different
//   module paths (github.com/dgrijalva/jwt-go and github.com/golang-jwt/jwt)
// TODO: Remove this replace statement after the bug below is resolved:
//       https://github.com/golang-migrate/migrate/issues/569
replace github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt v3.2.0+incompatible
