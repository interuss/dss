// This file is auto-generated; do not change as any changes will be overwritten
package main

import (
	"log"
	"net/http"
	"time"
)

type ExampleImplementation struct{}

func (*ExampleImplementation) QueryOperationalIntentReferences(body QueryOperationalIntentReferenceParameters, bodyParseError *error) QueryOperationalIntentReferencesResponseSet {
	response := QueryOperationalIntentReferencesResponseSet{}
	response.Response200 = &QueryOperationalIntentReferenceResponse{}
	return response
}

func (*ExampleImplementation) CreateOperationalIntentReference(entityid EntityID, body PutOperationalIntentReferenceParameters, bodyParseError *error) CreateOperationalIntentReferenceResponseSet {
	response := CreateOperationalIntentReferenceResponseSet{}
	response.Response201 = &ChangeOperationalIntentReferenceResponse{}
	return response
}

func (*ExampleImplementation) GetOperationalIntentReference(entityid EntityID) GetOperationalIntentReferenceResponseSet {
	response := GetOperationalIntentReferenceResponseSet{}
	response.Response200 = &GetOperationalIntentReferenceResponse{}
	return response
}

func (*ExampleImplementation) DeleteOperationalIntentReference(entityid EntityID, ovn EntityOVN) DeleteOperationalIntentReferenceResponseSet {
	response := DeleteOperationalIntentReferenceResponseSet{}
	response.Response200 = &ChangeOperationalIntentReferenceResponse{}
	return response
}

func (*ExampleImplementation) UpdateOperationalIntentReference(entityid EntityID, ovn EntityOVN, body PutOperationalIntentReferenceParameters, bodyParseError *error) UpdateOperationalIntentReferenceResponseSet {
	response := UpdateOperationalIntentReferenceResponseSet{}
	response.Response200 = &ChangeOperationalIntentReferenceResponse{}
	return response
}

func (*ExampleImplementation) QueryConstraintReferences(body QueryConstraintReferenceParameters, bodyParseError *error) QueryConstraintReferencesResponseSet {
	response := QueryConstraintReferencesResponseSet{}
	response.Response200 = &QueryConstraintReferencesResponse{}
	return response
}

func (*ExampleImplementation) CreateConstraintReference(entityid EntityID, body PutConstraintReferenceParameters, bodyParseError *error) CreateConstraintReferenceResponseSet {
	response := CreateConstraintReferenceResponseSet{}
	response.Response201 = &ChangeConstraintReferenceResponse{}
	return response
}

func (*ExampleImplementation) GetConstraintReference(entityid EntityID) GetConstraintReferenceResponseSet {
	response := GetConstraintReferenceResponseSet{}
	response.Response200 = &GetConstraintReferenceResponse{}
	return response
}

func (*ExampleImplementation) DeleteConstraintReference(entityid EntityID, ovn EntityOVN) DeleteConstraintReferenceResponseSet {
	response := DeleteConstraintReferenceResponseSet{}
	response.Response200 = &ChangeConstraintReferenceResponse{}
	return response
}

func (*ExampleImplementation) UpdateConstraintReference(entityid EntityID, ovn EntityOVN, body PutConstraintReferenceParameters, bodyParseError *error) UpdateConstraintReferenceResponseSet {
	response := UpdateConstraintReferenceResponseSet{}
	response.Response200 = &ChangeConstraintReferenceResponse{}
	return response
}

func (*ExampleImplementation) QuerySubscriptions(body QuerySubscriptionParameters, bodyParseError *error) QuerySubscriptionsResponseSet {
	response := QuerySubscriptionsResponseSet{}
	response.Response200 = &QuerySubscriptionsResponse{}
	return response
}

func (*ExampleImplementation) CreateSubscription(subscriptionid SubscriptionID, body PutSubscriptionParameters, bodyParseError *error) CreateSubscriptionResponseSet {
	response := CreateSubscriptionResponseSet{}
	response.Response200 = &PutSubscriptionResponse{}
	return response
}

func (*ExampleImplementation) GetSubscription(subscriptionid SubscriptionID) GetSubscriptionResponseSet {
	response := GetSubscriptionResponseSet{}
	response.Response200 = &GetSubscriptionResponse{}
	return response
}

func (*ExampleImplementation) DeleteSubscription(subscriptionid SubscriptionID, version string) DeleteSubscriptionResponseSet {
	response := DeleteSubscriptionResponseSet{}
	response.Response200 = &DeleteSubscriptionResponse{}
	return response
}

func (*ExampleImplementation) UpdateSubscription(subscriptionid SubscriptionID, version string, body PutSubscriptionParameters, bodyParseError *error) UpdateSubscriptionResponseSet {
	response := UpdateSubscriptionResponseSet{}
	response.Response200 = &PutSubscriptionResponse{}
	return response
}

func (*ExampleImplementation) MakeDssReport(body ErrorReport, bodyParseError *error) MakeDssReportResponseSet {
	response := MakeDssReportResponseSet{}
	response.Response201 = &ErrorReport{}
	return response
}

func (*ExampleImplementation) SetUssAvailability(uss_id string, body SetUssAvailabilityStatusParameters, bodyParseError *error) SetUssAvailabilityResponseSet {
	response := SetUssAvailabilityResponseSet{}
	response.Response200 = &UssAvailabilityStatusResponse{}
	return response
}

func (*ExampleImplementation) GetUssAvailability(uss_id string) GetUssAvailabilityResponseSet {
	response := GetUssAvailabilityResponseSet{}
	response.Response200 = &UssAvailabilityStatusResponse{}
	return response
}

func main() {
	router1 := MakeRouter(&ExampleImplementation{})
	multiRouter := MultiRouter{Routers: []*Router{&router1}}
	s := &http.Server{
		Addr:           ":8080",
		Handler:        &multiRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
