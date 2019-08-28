module github.com/steeling/InterUSS-Platform

// This forked version of openapi2proto has limited support for Open API v3.
replace github.com/NYTimes/openapi2proto => github.com/davidsansome/openapi2proto v0.2.3-0.20190826092301-b98d13b38dab

go 1.12

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/NYTimes/openapi2proto v0.2.2 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dolmen-go/jsonptr v0.0.0-20190605225012-a9a7ae01cd7d // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/golang/geo v0.0.0-20190507233405-a0e886e97a51
	github.com/golang/protobuf v1.3.2
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/pprof v0.0.0-20190723021845-34ac40c74b70 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/grpc-gateway v1.9.6
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/kr/pty v1.1.8 // indirect
	github.com/lib/pq v1.2.0
	github.com/rogpeppe/fastuuid v1.2.0 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/stretchr/testify v1.3.0
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586 // indirect
	golang.org/x/image v0.0.0-20190823064033-3a9bac650e44 // indirect
	golang.org/x/mobile v0.0.0-20190823173732-30c70e3810e9 // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/sys v0.0.0-20190825160603-fb81701db80f // indirect
	golang.org/x/tools v0.0.0-20190826060629-95c3470cfb70 // indirect
	google.golang.org/api v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.23.0
	honnef.co/go/tools v0.0.1-2019.2.2 // indirect
)
