// This file is auto-generated; do not change as any changes will be overwritten
package main

import (
	"example/api"
	"example/api/rid"
	"example/api/scd"
	"log"
	"net/http"
	"time"
)

type PermissiveAuthorizer struct{}

func (*PermissiveAuthorizer) Authorize(w http.ResponseWriter, r *http.Request, schemes *map[string]api.SecurityScheme) api.AuthorizationResult {
	return api.AuthorizationResult{}
}

type ScdImplementation struct{}

func (*ScdImplementation) QueryOperationalIntentReferences(req *scd.QueryOperationalIntentReferencesRequest) scd.QueryOperationalIntentReferencesResponseSet {
	response := scd.QueryOperationalIntentReferencesResponseSet{}
	response.Response200 = &scd.QueryOperationalIntentReferenceResponse{}
	return response
}

func (*ScdImplementation) GetOperationalIntentReference(req *scd.GetOperationalIntentReferenceRequest) scd.GetOperationalIntentReferenceResponseSet {
	response := scd.GetOperationalIntentReferenceResponseSet{}
	response.Response200 = &scd.GetOperationalIntentReferenceResponse{}
	return response
}

func (*ScdImplementation) CreateOperationalIntentReference(req *scd.CreateOperationalIntentReferenceRequest) scd.CreateOperationalIntentReferenceResponseSet {
	response := scd.CreateOperationalIntentReferenceResponseSet{}
	response.Response201 = &scd.ChangeOperationalIntentReferenceResponse{}
	return response
}

func (*ScdImplementation) UpdateOperationalIntentReference(req *scd.UpdateOperationalIntentReferenceRequest) scd.UpdateOperationalIntentReferenceResponseSet {
	response := scd.UpdateOperationalIntentReferenceResponseSet{}
	response.Response200 = &scd.ChangeOperationalIntentReferenceResponse{}
	return response
}

func (*ScdImplementation) DeleteOperationalIntentReference(req *scd.DeleteOperationalIntentReferenceRequest) scd.DeleteOperationalIntentReferenceResponseSet {
	response := scd.DeleteOperationalIntentReferenceResponseSet{}
	response.Response200 = &scd.ChangeOperationalIntentReferenceResponse{}
	return response
}

func (*ScdImplementation) QueryConstraintReferences(req *scd.QueryConstraintReferencesRequest) scd.QueryConstraintReferencesResponseSet {
	response := scd.QueryConstraintReferencesResponseSet{}
	response.Response200 = &scd.QueryConstraintReferencesResponse{}
	return response
}

func (*ScdImplementation) GetConstraintReference(req *scd.GetConstraintReferenceRequest) scd.GetConstraintReferenceResponseSet {
	response := scd.GetConstraintReferenceResponseSet{}
	response.Response200 = &scd.GetConstraintReferenceResponse{}
	return response
}

func (*ScdImplementation) CreateConstraintReference(req *scd.CreateConstraintReferenceRequest) scd.CreateConstraintReferenceResponseSet {
	response := scd.CreateConstraintReferenceResponseSet{}
	response.Response201 = &scd.ChangeConstraintReferenceResponse{}
	return response
}

func (*ScdImplementation) UpdateConstraintReference(req *scd.UpdateConstraintReferenceRequest) scd.UpdateConstraintReferenceResponseSet {
	response := scd.UpdateConstraintReferenceResponseSet{}
	response.Response200 = &scd.ChangeConstraintReferenceResponse{}
	return response
}

func (*ScdImplementation) DeleteConstraintReference(req *scd.DeleteConstraintReferenceRequest) scd.DeleteConstraintReferenceResponseSet {
	response := scd.DeleteConstraintReferenceResponseSet{}
	response.Response200 = &scd.ChangeConstraintReferenceResponse{}
	return response
}

func (*ScdImplementation) QuerySubscriptions(req *scd.QuerySubscriptionsRequest) scd.QuerySubscriptionsResponseSet {
	response := scd.QuerySubscriptionsResponseSet{}
	response.Response200 = &scd.QuerySubscriptionsResponse{}
	return response
}

func (*ScdImplementation) GetSubscription(req *scd.GetSubscriptionRequest) scd.GetSubscriptionResponseSet {
	response := scd.GetSubscriptionResponseSet{}
	response.Response200 = &scd.GetSubscriptionResponse{}
	return response
}

func (*ScdImplementation) CreateSubscription(req *scd.CreateSubscriptionRequest) scd.CreateSubscriptionResponseSet {
	response := scd.CreateSubscriptionResponseSet{}
	response.Response200 = &scd.PutSubscriptionResponse{}
	return response
}

func (*ScdImplementation) UpdateSubscription(req *scd.UpdateSubscriptionRequest) scd.UpdateSubscriptionResponseSet {
	response := scd.UpdateSubscriptionResponseSet{}
	response.Response200 = &scd.PutSubscriptionResponse{}
	return response
}

func (*ScdImplementation) DeleteSubscription(req *scd.DeleteSubscriptionRequest) scd.DeleteSubscriptionResponseSet {
	response := scd.DeleteSubscriptionResponseSet{}
	response.Response200 = &scd.DeleteSubscriptionResponse{}
	return response
}

func (*ScdImplementation) MakeDssReport(req *scd.MakeDssReportRequest) scd.MakeDssReportResponseSet {
	response := scd.MakeDssReportResponseSet{}
	response.Response201 = &scd.ErrorReport{}
	return response
}

func (*ScdImplementation) GetUssAvailability(req *scd.GetUssAvailabilityRequest) scd.GetUssAvailabilityResponseSet {
	response := scd.GetUssAvailabilityResponseSet{}
	response.Response200 = &scd.UssAvailabilityStatusResponse{}
	return response
}

func (*ScdImplementation) SetUssAvailability(req *scd.SetUssAvailabilityRequest) scd.SetUssAvailabilityResponseSet {
	response := scd.SetUssAvailabilityResponseSet{}
	response.Response200 = &scd.UssAvailabilityStatusResponse{}
	return response
}

type RidImplementation struct{}

func (*RidImplementation) SearchIdentificationServiceAreas(req *rid.SearchIdentificationServiceAreasRequest) rid.SearchIdentificationServiceAreasResponseSet {
	response := rid.SearchIdentificationServiceAreasResponseSet{}
	response.Response200 = &rid.SearchIdentificationServiceAreasResponse{}
	return response
}

func (*RidImplementation) GetIdentificationServiceArea(req *rid.GetIdentificationServiceAreaRequest) rid.GetIdentificationServiceAreaResponseSet {
	response := rid.GetIdentificationServiceAreaResponseSet{}
	response.Response200 = &rid.GetIdentificationServiceAreaResponse{}
	return response
}

func (*RidImplementation) CreateIdentificationServiceArea(req *rid.CreateIdentificationServiceAreaRequest) rid.CreateIdentificationServiceAreaResponseSet {
	response := rid.CreateIdentificationServiceAreaResponseSet{}
	response.Response200 = &rid.PutIdentificationServiceAreaResponse{}
	return response
}

func (*RidImplementation) UpdateIdentificationServiceArea(req *rid.UpdateIdentificationServiceAreaRequest) rid.UpdateIdentificationServiceAreaResponseSet {
	response := rid.UpdateIdentificationServiceAreaResponseSet{}
	response.Response200 = &rid.PutIdentificationServiceAreaResponse{}
	return response
}

func (*RidImplementation) DeleteIdentificationServiceArea(req *rid.DeleteIdentificationServiceAreaRequest) rid.DeleteIdentificationServiceAreaResponseSet {
	response := rid.DeleteIdentificationServiceAreaResponseSet{}
	response.Response200 = &rid.DeleteIdentificationServiceAreaResponse{}
	return response
}

func (*RidImplementation) SearchSubscriptions(req *rid.SearchSubscriptionsRequest) rid.SearchSubscriptionsResponseSet {
	response := rid.SearchSubscriptionsResponseSet{}
	response.Response200 = &rid.SearchSubscriptionsResponse{}
	return response
}

func (*RidImplementation) GetSubscription(req *rid.GetSubscriptionRequest) rid.GetSubscriptionResponseSet {
	response := rid.GetSubscriptionResponseSet{}
	response.Response200 = &rid.GetSubscriptionResponse{}
	return response
}

func (*RidImplementation) CreateSubscription(req *rid.CreateSubscriptionRequest) rid.CreateSubscriptionResponseSet {
	response := rid.CreateSubscriptionResponseSet{}
	response.Response200 = &rid.PutSubscriptionResponse{}
	return response
}

func (*RidImplementation) UpdateSubscription(req *rid.UpdateSubscriptionRequest) rid.UpdateSubscriptionResponseSet {
	response := rid.UpdateSubscriptionResponseSet{}
	response.Response200 = &rid.PutSubscriptionResponse{}
	return response
}

func (*RidImplementation) DeleteSubscription(req *rid.DeleteSubscriptionRequest) rid.DeleteSubscriptionResponseSet {
	response := rid.DeleteSubscriptionResponseSet{}
	response.Response200 = &rid.DeleteSubscriptionResponse{}
	return response
}

func main() {
	authorizer := PermissiveAuthorizer{}
	scdRouter := scd.MakeAPIRouter(&ScdImplementation{}, &authorizer)
	ridRouter := rid.MakeAPIRouter(&RidImplementation{}, &authorizer)
	multiRouter := api.MultiRouter{Routers: []api.APIRouter{&scdRouter, &ridRouter}}
	s := &http.Server{
		Addr:           ":8080",
		Handler:        &multiRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
