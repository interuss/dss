module github.com/interuss/dss/docker/generator

go 1.14

replace github.com/NYTimes/openapi2proto => github.com/davidsansome/openapi2proto v0.2.3-0.20190826092301-b98d13b38dab

require (
	github.com/golang/protobuf v1.4.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.3
	google.golang.org/genproto v0.0.0-20200519141106-08726f379972
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.22.0
)