from monitoring.uss_qualifier.scd.data_interfaces import KnownIssueFields


nominal_test_common_error_notification = KnownIssueFields(test_code = "nominal_test", 
                                                relevant_requirements = [], 
                                                severity = "High",
                                                subject="", 
                                                summary ="Injection request for a valid flight was unsuccessful", details = "All operational intent data provided was complete and correct with no airspace conflicts, conforming to the relevant standardized formats and the data should have been processed successfully and flight should have been planned.")        
        
if_planned_with_conflict_with_flight_explanation = KnownIssueFields(test_code = "nominal_test", 
                                            relevant_requirements = ["A operational intent that has time or space conflict should not be planned by the USS"], severity = "High", 
                                            subject="Operational Intent provided should not be sucessfully planned by the USSP", 
                                            summary ="The operational intent details provided were generated in such a way that they should not have been planned.", details = "The co-ordinates of the 4D Operational intent conflicts with an existing operational intent in the area and the processing result should not be a successful planning of the intent.")
                                            
common_conflict_with_flight_explanation = KnownIssueFields(test_code = "nominal_test", 
                                                            relevant_requirements = ["A complete and correct flight authorisation data should be provided by the USSP."], 
                                                            severity = "High", 
                                                            subject="Flight authorisation data is incorrect and operational intent should not be processed", 
                                                            summary ="Invalid operational intent data was provided and therefore the operational intent should not have been planned or submitted to the DSS", 
                                                            details = "All operational intent data should be validated by the USSP before submitting the Operational Intent to the DSS. In this case, the data is not valid, processing should have returned a error.")
        

flight_authorisation_test_common_error_notification = KnownIssueFields(test_code = "flight_authorisation_test", 
                                                relevant_requirements = [], 
                                                severity = "High",
                                                subject="", 
                                                summary ="Flight authorisation request for with valid flight details should be processed successfully", details = "All data provided was complete and correct with no errors, conforming to the relevant standardized formats and the data should have been processed successfully.")        
        
if_conflict_with_flight_explanation = KnownIssueFields(test_code = "nominal_test", 
                                                        relevant_requirements = ["A operational intent that has no time or space conflict should be planned by the USS"], 
                                                        severity= "High", 
                                                        subject="Operational Intent provided should be planned successfully", 
                                                        summary ="The operational intent details provided were generated in such a way that they should have been planned.", 
                                                        details = "The co-ordinates of the 4D Operational intent does not conflict with any existing operational intents in the area and the processing result should be a successful planning of the intent.")             

if_planned_with_incorrect_uas_serial_number_explanation = KnownIssueFields(test_code = "flight_authorisation_test",
                                            relevant_requirements = ["ANNEX IV of Commission Implementing Regulation (EU) 2021/664, paragraph 1"], 
                                            severity = "High", 
                                            subject="UAS Serial Number provided is incorrect", 
                                            summary ="Flight created with invalid UAS serial number", 
                                            details = "The UAS serial number is not as expressed in the ANSI/CTA-2063 Physical Serial Number format and should be rejected.")   


if_planned_with_incorrect_operator_registration_number_explanation = KnownIssueFields(test_code = "flight_authorisation_test", 
                                            relevant_requirements = ["ANNEX IV of COMMISSION IMPLEMENTING REGULATION (EU) 2021/664, paragraph 1"], 
                                            severity = "High", 
                                            subject="Operator Registration Number provided is incorrect", 
                                            summary ="Flight created with invalid Operator registration number", 
                                            details = "The UAS serial number provided is not as expressed as described in the EN4709-02 standard should be rejected.")

