// This file is auto-generated; do not change as any changes will be overwritten
package main

import (
	"log"
	"net/http"
	"time"
)

type PermissiveAuthorizer struct{}

func (*PermissiveAuthorizer) Authorize(w http.ResponseWriter, r *http.Request, schemes *map[string]SecurityScheme) AuthorizationResult {
	return AuthorizationResult{}
}

type ExampleImplementation struct{}

func (*ExampleImplementation) QueryOperationalIntentReferences(req *QueryOperationalIntentReferencesRequest) QueryOperationalIntentReferencesResponseSet {
	response := QueryOperationalIntentReferencesResponseSet{}
	response.Response200 = &QueryOperationalIntentReferenceResponse{}
	return response
}

func (*ExampleImplementation) GetOperationalIntentReference(req *GetOperationalIntentReferenceRequest) GetOperationalIntentReferenceResponseSet {
	response := GetOperationalIntentReferenceResponseSet{}
	response.Response200 = &GetOperationalIntentReferenceResponse{}
	return response
}

func (*ExampleImplementation) CreateOperationalIntentReference(req *CreateOperationalIntentReferenceRequest) CreateOperationalIntentReferenceResponseSet {
	response := CreateOperationalIntentReferenceResponseSet{}
	response.Response201 = &ChangeOperationalIntentReferenceResponse{}
	return response
}

func (*ExampleImplementation) UpdateOperationalIntentReference(req *UpdateOperationalIntentReferenceRequest) UpdateOperationalIntentReferenceResponseSet {
	response := UpdateOperationalIntentReferenceResponseSet{}
	response.Response200 = &ChangeOperationalIntentReferenceResponse{}
	return response
}

func (*ExampleImplementation) DeleteOperationalIntentReference(req *DeleteOperationalIntentReferenceRequest) DeleteOperationalIntentReferenceResponseSet {
	response := DeleteOperationalIntentReferenceResponseSet{}
	response.Response200 = &ChangeOperationalIntentReferenceResponse{}
	return response
}

func (*ExampleImplementation) QueryConstraintReferences(req *QueryConstraintReferencesRequest) QueryConstraintReferencesResponseSet {
	response := QueryConstraintReferencesResponseSet{}
	response.Response200 = &QueryConstraintReferencesResponse{}
	return response
}

func (*ExampleImplementation) GetConstraintReference(req *GetConstraintReferenceRequest) GetConstraintReferenceResponseSet {
	response := GetConstraintReferenceResponseSet{}
	response.Response200 = &GetConstraintReferenceResponse{}
	return response
}

func (*ExampleImplementation) CreateConstraintReference(req *CreateConstraintReferenceRequest) CreateConstraintReferenceResponseSet {
	response := CreateConstraintReferenceResponseSet{}
	response.Response201 = &ChangeConstraintReferenceResponse{}
	return response
}

func (*ExampleImplementation) UpdateConstraintReference(req *UpdateConstraintReferenceRequest) UpdateConstraintReferenceResponseSet {
	response := UpdateConstraintReferenceResponseSet{}
	response.Response200 = &ChangeConstraintReferenceResponse{}
	return response
}

func (*ExampleImplementation) DeleteConstraintReference(req *DeleteConstraintReferenceRequest) DeleteConstraintReferenceResponseSet {
	response := DeleteConstraintReferenceResponseSet{}
	response.Response200 = &ChangeConstraintReferenceResponse{}
	return response
}

func (*ExampleImplementation) QuerySubscriptions(req *QuerySubscriptionsRequest) QuerySubscriptionsResponseSet {
	response := QuerySubscriptionsResponseSet{}
	response.Response200 = &QuerySubscriptionsResponse{}
	return response
}

func (*ExampleImplementation) GetSubscription(req *GetSubscriptionRequest) GetSubscriptionResponseSet {
	response := GetSubscriptionResponseSet{}
	response.Response200 = &GetSubscriptionResponse{}
	return response
}

func (*ExampleImplementation) CreateSubscription(req *CreateSubscriptionRequest) CreateSubscriptionResponseSet {
	response := CreateSubscriptionResponseSet{}
	response.Response200 = &PutSubscriptionResponse{}
	return response
}

func (*ExampleImplementation) UpdateSubscription(req *UpdateSubscriptionRequest) UpdateSubscriptionResponseSet {
	response := UpdateSubscriptionResponseSet{}
	response.Response200 = &PutSubscriptionResponse{}
	return response
}

func (*ExampleImplementation) DeleteSubscription(req *DeleteSubscriptionRequest) DeleteSubscriptionResponseSet {
	response := DeleteSubscriptionResponseSet{}
	response.Response200 = &DeleteSubscriptionResponse{}
	return response
}

func (*ExampleImplementation) MakeDssReport(req *MakeDssReportRequest) MakeDssReportResponseSet {
	response := MakeDssReportResponseSet{}
	response.Response201 = &ErrorReport{}
	return response
}

func (*ExampleImplementation) GetUssAvailability(req *GetUssAvailabilityRequest) GetUssAvailabilityResponseSet {
	response := GetUssAvailabilityResponseSet{}
	response.Response200 = &UssAvailabilityStatusResponse{}
	return response
}

func (*ExampleImplementation) SetUssAvailability(req *SetUssAvailabilityRequest) SetUssAvailabilityResponseSet {
	response := SetUssAvailabilityResponseSet{}
	response.Response200 = &UssAvailabilityStatusResponse{}
	return response
}

func main() {
	router1 := MakeAPIRouter(&ExampleImplementation{}, &PermissiveAuthorizer{})
	multiRouter := MultiRouter{Routers: []*APIRouter{&router1}}
	s := &http.Server{
		Addr:           ":8080",
		Handler:        &multiRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
