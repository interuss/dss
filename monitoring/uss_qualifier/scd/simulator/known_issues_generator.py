from monitoring.uss_qualifier.scd.data_interfaces import KnownIssueFields

nominal_planning_test_subject = "Operational Intent Processsing"

nominal_planning_test_common_error_notification = KnownIssueFields(test_code = "nominal_planning_test", 
                                                relevant_requirements = [], 
                                                severity = "High",
                                                subject= nominal_planning_test_subject, 
                                                summary ="Injection request for a valid flight was unsuccessful", details = "All operational intent and flight authorisation data provided was complete and correct with no airspace conflicts. The operational intent data should have been processed successfully and flight should have been planned.")        
        
nominal_planning_test_rejected_error_notification = KnownIssueFields(test_code = "nominal_planning_test", 
                                                relevant_requirements = [], 
                                                severity = "High",
                                                subject=nominal_planning_test_subject, 
                                                summary ="Injection request for a valid flight was rejected", details = "All operational intent and flight authorisation data provided was complete and correct with no airspace conflicts. The operational intent data should have been processed successfully and flight should have been planned.")        
        
nominal_planning_test_failed_error_notification = KnownIssueFields(test_code = "nominal_planning_test", 
                                                relevant_requirements = [], 
                                                severity = "High",
                                                subject="", 
                                                summary ="Injection request for a valid flight failed", details = "All operational intent and flight authorisation data provided was complete and correct with no airspace conflicts. The operational intent data should have been processed successfully and flight should have been planned.")        
        
if_planned_with_conflict_with_flight_explanation = KnownIssueFields(test_code = "nominal_planning_test", 
                                            relevant_requirements = [], 
                                            severity = "High", 
                                            subject=nominal_planning_test_subject, 
                                            summary ="The operational intent details provided were generated in such a way that they should not have been planned.", details = "The co-ordinates of the 4D Operational intent conflicts with an existing operational intent in the area and the processing result should not be a successful planning of the intent.")
                                            
conflict_with_flight_explanation = KnownIssueFields(test_code = "nominal_planning_test", 
                                                            relevant_requirements = [], 
                                                            severity = "High", 
                                                            subject=nominal_planning_test_subject, 
                                                            summary ="The operational intent data provided should have been processed without conflicts", 
                                                            details = "All operational intent data provided is correct and valid and free of conflict in space and time, therefore it should have been planned by the USSP.")
        

if_conflict_with_flight_explanation = KnownIssueFields(test_code = "nominal_planning_test", 
                                                        relevant_requirements = [], 
                                                        severity= "High", 
                                                        subject=nominal_planning_test_subject, 
                                                        summary ="The operational intent details provided were generated in such a way that they should have been planned.", 
                                                        details = "The co-ordinates of the 4D Operational intent does not conflict with any existing operational intents in the area and the processing result should be a successful planning of the intent.")             

flight_authorisation_test_common_error_notification = KnownIssueFields(test_code = "flight_authorisation_validation_test", 
                                                relevant_requirements = [], 
                                                severity = "High",
                                                subject="Flight Authorisation Data", 
                                                summary ="Flight authorisation request with valid flight details should be processed successfully", details = "All data provided was complete and correct with no errors, conforming to the relevant standardized formats and the data should have been processed successfully.")        
        

flight_authorisation_test_conflict_with_flight_error_notification = KnownIssueFields(test_code = "flight_authorisation_validation_test", 
                                                relevant_requirements = [], 
                                                severity = "High",
                                                subject="Flight Authorisation Data", 
                                                summary ="Flight authorisation request did not contain any operational intents and therefore should not lead to a airspace conflict error.", details = "Operational intents are provided for nominal tests only and flight planning is not expected for flight authorisation test.")        
        
if_planned_with_incorrect_uas_serial_number_explanation = KnownIssueFields(test_code = "flight_authorisation_validation_test",
                                            relevant_requirements = ["ANNEX IV of Commission Implementing Regulation (EU) 2021/664, paragraph 1"], 
                                            severity = "High", 
                                            subject="UAS Serial Number", 
                                            summary ="Flight created with invalid UAS serial number", 
                                            details = "The UAS serial number is not as expressed in the ANSI/CTA-2063 Physical Serial Number format and should be rejected.")   


if_planned_with_incorrect_operator_registration_number_explanation = KnownIssueFields(test_code = "flight_authorisation_validation_test", 
                                            relevant_requirements = ["ANNEX IV of COMMISSION IMPLEMENTING REGULATION (EU) 2021/664, paragraph 1"], 
                                            severity = "High", 
                                            subject="Operator Registration ID", 
                                            summary ="Flight created with invalid Operator registration ID", 
                                            details = "The UAS serial number provided is not as expressed as described in the EN4709-02 standard should be rejected.")

