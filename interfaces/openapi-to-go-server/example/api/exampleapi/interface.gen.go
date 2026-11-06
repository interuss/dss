// This file is auto-generated; do not change as any changes will be overwritten
package exampleapi

import (
	"context"
	"example/api"
)

var (
	GetGreetingSecurity = []api.AuthorizationOption{}
)

type GetGreetingRequest struct {
	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetGreetingOperationID = "exampleapi.GetGreeting"

func (request *GetGreetingRequest) OperationID() string { return GetGreetingOperationID }

type GetGreetingResponseSet struct {
	// A greeting
	Response200 *Greeting

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type Implementation interface {
	GetGreeting(ctx context.Context, req *GetGreetingRequest) GetGreetingResponseSet
}
