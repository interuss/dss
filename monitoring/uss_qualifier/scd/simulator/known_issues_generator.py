from monitoring.uss_qualifier.scd.data_interfaces import KnownIssueFields
from typing import Dict
### Begin Nominal planning test notifications ###


class NominalTestKnownIssuesAcceptableResults():
    """A class to generate Known Issues and Acceptable results for the nominal test data"""

    def __init__(self, expected_flight_authorisation_processing_result:str, expected_operational_intent_processing_result:str):

        self.expected_flight_authorisation_processing_result = expected_flight_authorisation_processing_result
        self.expected_operational_intent_processing_result = expected_operational_intent_processing_result

        self.nominal_planning_test_subject = "Operational Intent Processsing"
        self.nominal_planning_test_if_conflict_with_flight_notification = KnownIssueFields(test_code = "nominal_planning_test", 
                                                                relevant_requirements = [], 
                                                                severity= "High", 
                                                                subject= self.nominal_planning_test_subject, 
                                                                summary ="The operational intent details provided were generated in such a way that they should have been planned.", 
                                                                details = "The co-ordinates of the 4D Operational intent does not conflict with any existing operational intents in the area and the processing result should be a successful planning of the intent.")             

        self.nominal_planning_test_common_error_notification = KnownIssueFields(test_code = "nominal_planning_test", 
                                                        relevant_requirements = [], 
                                                        severity = "High",
                                                        subject= self.nominal_planning_test_subject, 
                                                        summary ="Injection request for a valid flight was unsuccessful", 
                                                        details = "All operational intent and flight authorisation data provided was complete and correct with no airspace conflicts. The operational intent data should have been processed successfully and flight should have been planned.")        
                
        self.nominal_planning_test_rejected_error_notification = KnownIssueFields(test_code = "nominal_planning_test", 
                                                        relevant_requirements = [], 
                                                        severity = "High",
                                                        subject= self.nominal_planning_test_subject, 
                                                        summary ="Injection request for a valid flight was rejected", 
                                                        details = "All operational intent and flight authorisation data provided was complete and correct with no airspace conflicts. The operational intent data should have been processed successfully and flight should have been planned.")        
                
        self.nominal_planning_test_failed_error_notification = KnownIssueFields(test_code = "nominal_planning_test", 
                                                        relevant_requirements = [], 
                                                        severity = "High",
                                                        subject= self.nominal_planning_test_subject, 
                                                        summary ="Injection request for a valid flight failed", 
                                                        details = "All operational intent and flight authorisation data provided was complete and correct with no airspace conflicts. The operational intent data should have been processed successfully and flight should have been planned.")        
                
        self.nominal_planning_test_if_planned_with_conflict_with_flight_notification = KnownIssueFields(test_code = "nominal_planning_test", 
                                                    relevant_requirements = [], 
                                                    severity = "High", 
                                                    subject= self.nominal_planning_test_subject, 
                                                    summary ="The operational intent details provided were generated in such a way that they should not have been planned.", 
                                                    details = "The co-ordinates of the 4D Operational intent conflicts with an existing operational intent in the area and the processing result should not be a successful planning of the intent.")
                                                    
        self.nominal_planning_test_conflict_with_flight_notification = KnownIssueFields(test_code = "nominal_planning_test", 
                                                                    relevant_requirements = [], 
                                                                    severity = "High", 
                                                                    subject= self.nominal_planning_test_subject, 
                                                                    summary ="The operational intent data provided should have been processed without conflicts", 
                                                                    details = "All operational intent data provided is correct and valid and free of conflict in space and time, therefore it should have been planned by the USSP.")


    def generate_nominal_test_known_issues_fields(self)-> Dict[str, KnownIssueFields]:
        """A method to generate messages for the user to take remedial actions when a nominal test returns a status that is not expected """
        all_known_issues_fields = {}
        
        if self.expected_operational_intent_processing_result == "Planned":
            all_known_issues_fields['ConflictWithFlight']= self.nominal_planning_test_if_conflict_with_flight_notification
            all_known_issues_fields['Rejected']= self.nominal_planning_test_rejected_error_notification
            all_known_issues_fields['Failed']= self.nominal_planning_test_failed_error_notification
        elif self.expected_operational_intent_processing_result == "ConflictWithFlight":
            all_known_issues_fields['Rejected']= self.nominal_planning_test_common_error_notification
            all_known_issues_fields['Failed']= self.nominal_planning_test_common_error_notification
            all_known_issues_fields['Planned']= self.nominal_planning_test_if_planned_with_conflict_with_flight_notification

        return all_known_issues_fields




### End Nominal planning test notifications ###                                                            

### Begin Nominal planning test (with priority) notifications ###                                                            

class NominalTestwPrioritiesKnownIssuesAcceptableResults():
    """A class to generate Known Issues and Acceptable results to inform about  data processing errors in the test data provided """

    def __init__(self, expected_flight_authorisation_processing_result:str, expected_operational_intent_processing_result:str):

        self.expected_flight_authorisation_processing_result = expected_flight_authorisation_processing_result
        self.expected_operational_intent_processing_result = expected_operational_intent_processing_result

        self.nominal_planning_with_priority_test_subject =  "Priority Operational Intent Processsing"
        self.nominal_planning_test_with_priority_conflict_with_flight_notification = KnownIssueFields(test_code = "nominal_planning_test_with_priority", 
                                                                    relevant_requirements = [], 
                                                                    severity = "High", 
                                                                    subject =  self.nominal_planning_with_priority_test_subject,
                                                                    summary ="The operational intent data provided should have been planned regardless of an airspace or time conflict", 
                                                                    details = "All operational intent data provided is correct and valid and while it has conflicts with an existing operational intent, the priority is higher so the USSP should have planned it.")
                                                                    
                                                                    
        self.nominal_planning_test_with_priority_rejected_error_notification = KnownIssueFields(test_code = "nominal_planning_test_with_priority", 
                                                        relevant_requirements = [], 
                                                        severity = "High",
                                                        subject = self.nominal_planning_with_priority_test_subject,
                                                        summary ="Injection request for a valid flight was rejected", 
                                                        details = "All operational intent and flight authorisation data provided was complete and correct. The operational intent data should have been processed successfully and flight should have been planned.")        
                
        self.nominal_planning_test_with_priority_failed_error_notification = KnownIssueFields(test_code = "nominal_planning_test_with_priority", 
                                                        relevant_requirements = [], 
                                                        severity = "High",
                                                        subject=self.nominal_planning_with_priority_test_subject, 
                                                        summary ="Injection request for a valid flight failed", 
                                                        details = "All operational intent and flight authorisation data provided was complete and correct. The operational intent data should have been processed successfully and flight should have been planned.")        
                
    def generate_nominal_test_with_priroties_known_issues_fields(self)-> Dict[str, KnownIssueFields]:
        """A method to generate messages for the user to take remedial actions when a nominal test with priorities with different priorties returns a status that is not expected """
        all_known_issues_fields = {}

        if self.expected_operational_intent_processing_result == "Planned":
            all_known_issues_fields['ConflictWithFlight']= self.nominal_planning_test_with_priority_conflict_with_flight_notification
            all_known_issues_fields['Rejected']= self.nominal_planning_test_with_priority_rejected_error_notification
            all_known_issues_fields['Failed']= self.nominal_planning_test_with_priority_failed_error_notification
        return all_known_issues_fields



### End Nominal planning test (with priority) notifications ###      

### Begin flight authorisation data validation notifications ###              

class FlightAuthorisationKnownIssuesAcceptableResults():
    """A class to generate Known Issues and Acceptable results as a result of flight authorisation data processing"""

    def __init__(self, expected_flight_authorisation_processing_result:str, expected_operational_intent_processing_result:str):

        self.expected_flight_authorisation_processing_result = expected_flight_authorisation_processing_result
        self.expected_operational_intent_processing_result = expected_operational_intent_processing_result

        self.flight_authorisation_test_subject = "Flight Authorisation Data"
        self.flight_authorisation_test_conflict_with_flight_error_notification = KnownIssueFields(test_code = "flight_authorisation_validation_test", 
                                                        relevant_requirements = [], 
                                                        severity = "High",
                                                        subject= "Operational Intent Processsing",
                                                        summary ="Flight authorisation request contains operational intents with no conflict in space and time and therefore should not lead to a airspace conflict error.", 
                                                        details = "The test data contains operational intents with no conflicts in space and time and therefore should be planned successfully.")        

        self.flight_authorisation_test_failed_with_without_incorrect_field_notification = KnownIssueFields(test_code = "flight_authorisation_validation_test", 
                                                        relevant_requirements = [], 
                                                        severity = "High",
                                                        subject= self.flight_authorisation_test_subject, 
                                                        summary ="Flight injection request contains flight authorisation with all required fields and a valid operational intent.", details = "The test data contains operational intents and flight authorisation data that are complete and valid and should be processsed by the USSP.")
                
        self.if_planned_with_incorrect_uas_serial_number_notification = KnownIssueFields(test_code = "flight_authorisation_validation_test",
                                                    relevant_requirements = ["ANNEX IV of Commission Implementing Regulation (EU) 2021/664, paragraph 1"], 
                                                    severity = "High", 
                                                    subject= "UAS Serial Number", 
                                                    summary ="Flight created with invalid UAS serial number", 
                                                    details = "The UAS serial number provided in the flight authorisation datais was not as expressed in the ANSI/CTA-2063 Physical Serial Number format and should be rejected.")   

        self.if_planned_with_incorrect_operator_registration_number_notification = KnownIssueFields(test_code = "flight_authorisation_validation_test", 
                                                    relevant_requirements = ["ANNEX IV of COMMISSION IMPLEMENTING REGULATION (EU) 2021/664, paragraph 1"], 
                                                    severity = "High", 
                                                    subject= "Operator Registration ID", 
                                                    summary ="Flight created with invalid Operator registration ID", 
                                                    details = "The Operation Registration ID provided in the flight authorisation details is not as expressed as described in the EN4709-02 standard and should be rejected.")

        self.flight_authorisation_test_common_error_notification = KnownIssueFields(test_code = "flight_authorisation_validation_test", 
                                                        relevant_requirements = [], 
                                                        severity = "High",
                                                        subject= self.flight_authorisation_test_subject, 
                                                        summary ="Flight authorisation request with valid flight details should be processed successfully", 
                                                        details = "All data provided was complete and correct with no errors, conforming to the relevant standardized formats and the data should have been processed successfully.")    

    def generate_flight_authorisation_test_known_issue_fields(self, incorrect_field:str = None)-> Dict[str, KnownIssueFields]:
        """A method to generate messages for the user to take remedial actions when a flight authorisation test returns a status that is not expected """

        all_known_issues_fields = {}
        all_known_issues_fields["Failed"] = self.flight_authorisation_test_failed_with_without_incorrect_field_notification            
        all_known_issues_fields["ConflictWithFlight"] = self.flight_authorisation_test_conflict_with_flight_error_notification

        if self.expected_flight_authorisation_processing_result == "Rejected":
            if incorrect_field == "uas_serial_number":
                all_known_issues_fields["Planned"] = self.if_planned_with_incorrect_uas_serial_number_notification
            elif incorrect_field == "operator_registration_number":
                all_known_issues_fields["Planned"] = self.if_planned_with_incorrect_operator_registration_number_notification
            
        elif self.expected_flight_authorisation_processing_result == "Planned":
            all_known_issues_fields["Rejected"] = self.flight_authorisation_test_common_error_notification

        return all_known_issues_fields
   

### End flight authorisation data validation notifications ###