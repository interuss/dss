from typing import List, Dict
from monitoring.uss_qualifier.scd.data_interfaces import FlightInjectionAttempt, InjectionTarget, KnownResponses, AutomatedTest, TestStep, Capability, RequiredUSSCapabilities, KnownIssueFields
from monitoring.monitorlib.scd import Time, Volume3D, Volume4D, Polygon, Altitude, LatLngPoint
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import OperationalIntentTestInjection,FlightAuthorisationData, InjectFlightRequest, UASClass

def parse_and_load_existing_test(current_test_definition)-> AutomatedTest:
    """A method to parse and load existing test definition """    
    raw_test_steps = current_test_definition['steps']
    current_capabilities = current_test_definition['uss_capabilities']

    all_test_steps: List[TestStep] = []
    all_uss_capabilities: List[RequiredUSSCapabilities] = []

    for uss_capability in current_capabilities:
        raw_capabilities = uss_capability['capabilities']
        raw_generate_issue = uss_capability['generate_issue'] if 'generate_issue' in uss_capability else None
        raw_injection_target = uss_capability['injection_target']
        all_capabilities = []
        for raw_capability in raw_capabilities: 
            all_capabilities.append(Capability(raw_capability))

        injection_target = InjectionTarget(uss_role=raw_injection_target['uss_role'])
        if raw_generate_issue:
            generate_issue = KnownIssueFields(test_code =raw_generate_issue['test_code'], relevant_requirements =raw_generate_issue['relevant_requirements'] ,severity=raw_generate_issue['severity'] ,  subject=raw_generate_issue['subject'] ,summary=raw_generate_issue['summary'], details= raw_generate_issue['details'])
        else: 
            generate_issue = None

        required_uss_capabilities = RequiredUSSCapabilities(capabilities = all_capabilities,injection_target = injection_target,  generate_issue = generate_issue)

        all_uss_capabilities.append(required_uss_capabilities)

    for idx, step_details in enumerate(raw_test_steps):
        # print(step_details.keys())
        inject_flight_details = step_details['inject_flight']
        
        test_injection_op_details = inject_flight_details['test_injection']['operational_intent']
        test_injection_flight_auth_details = inject_flight_details['test_injection']['flight_authorisation']
        all_volumes_raw = test_injection_op_details['volumes']
        all_4d_volumes: List[Volume4D] = []
        for volume in all_volumes_raw:
            altitude_lower = Altitude(value = volume['volume']['altitude_lower']['value'] , reference = volume['volume']['altitude_lower']['reference'], units =volume['volume']['altitude_lower']['units']  )
            altitude_upper =  Altitude(value = volume['volume']['altitude_upper']['value'] , reference = volume['volume']['altitude_upper']['reference'], units =volume['volume']['altitude_upper']['units']  )

            if 'outline_polygon' in volume['volume'].keys():
                all_vertices = volume['volume']['outline_polygon']['vertices']
                polygon_verticies = []
                for vertex in all_vertices:
                    v = LatLngPoint(lat = vertex['lat'],lng=vertex['lng'])
                    polygon_verticies.append(v)

                outline_polygon = Polygon(vertices=polygon_verticies)
            else: 
                outline_polygon = None
            current_3d_volume = Volume3D(outline_polygon=outline_polygon, altitude_upper=altitude_upper, altitude_lower= altitude_lower )
            
            time_start = Time(value = volume['time_start']['value'], format = volume['time_start']['format'])
            time_end = Time(value = volume['time_end']['value'], format = volume['time_end']['format'])
            
            current_4d_volume = Volume4D(volume = current_3d_volume, time_start=time_start , time_end =time_end)

            all_4d_volumes.append(current_4d_volume)

        if 'all_off_nominal_volumes' in volume.keys():
            all_off_nominal_volumes = test_injection_op_details['all_off_nominal_volumes']
        else: 
            all_off_nominal_volumes = []

        operational_intent_test_injection = OperationalIntentTestInjection(state = test_injection_op_details['state'], priority = test_injection_op_details['priority'], volumes = all_4d_volumes, off_nominal_volumes = all_off_nominal_volumes)

        uas_type_certificate = test_injection_flight_auth_details['uas_type_certificate'] if 'uas_type_certificate' in test_injection_flight_auth_details else ''
        uas_id = test_injection_flight_auth_details['uas_id'] if 'uas_id' in test_injection_flight_auth_details else ''


        flight_authorisation_data = FlightAuthorisationData(uas_serial_number = test_injection_flight_auth_details['uas_serial_number'], operation_mode = test_injection_flight_auth_details['operation_mode'],operation_category = test_injection_flight_auth_details['operation_category'], uas_class=UASClass(test_injection_flight_auth_details['uas_class']),  identification_technologies=test_injection_flight_auth_details['identification_technologies'],uas_type_certificate=uas_type_certificate,  connectivity_methods =test_injection_flight_auth_details['connectivity_methods'],endurance_minutes=test_injection_flight_auth_details['endurance_minutes'], emergency_procedure_url=test_injection_flight_auth_details['emergency_procedure_url'],operator_id=test_injection_flight_auth_details['operator_id'],  uas_id=uas_id)

        test_injection = InjectFlightRequest(operational_intent = operational_intent_test_injection, flight_authorisation = flight_authorisation_data)

        incorrect_result_details: Dict[str, KnownIssueFields] = {}
        for k,v in inject_flight_details['known_responses']['incorrect_result_details'].items():
            incorrect_result_details[k] = KnownIssueFields(test_code = v['test_code'],relevant_requirements = v['relevant_requirements'],severity = v['severity'],  subject = v['subject'], summary = v['summary'],details = v['details'])
        
        known_responses  = KnownResponses(acceptable_results = inject_flight_details['known_responses']['acceptable_results'], incorrect_result_details = incorrect_result_details)
        
        inj_target = InjectionTarget(uss_role=inject_flight_details['injection_target']['uss_role'])

        inject_flight = FlightInjectionAttempt(reference_time = inject_flight_details['reference_time'], name = inject_flight_details['name'], test_injection = test_injection, known_responses = known_responses, injection_target = inj_target)
        
        current_step = TestStep(name =step_details['name'], inject_flight = inject_flight, delete_flight = [])
        all_test_steps.append(current_step)
            
    # over write the steps with the new steps
            
    current_test = AutomatedTest(name = current_test_definition['name'], steps = all_test_steps, uss_capabilities = all_uss_capabilities)

    return current_test
    