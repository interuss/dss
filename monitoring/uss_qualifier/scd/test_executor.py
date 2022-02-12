from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest
from monitoring.uss_qualifier.rid.utils import InjectionTargetConfiguration
from monitoring.uss_qualifier.scd.data_interfaces import AutomatedTest, TestStep, FlightInjectionAttempt, \
    InjectionTarget, FlightDeletionAttempt, KnownResponses
from monitoring.uss_qualifier.scd.executor import TestRunner, targets_combination


targets = [
    InjectionTargetConfiguration(
        name="uss_unit_test_1",
        injection_base_url="http://host.docker.internal:8074/scdsc"
    )
]

automated_test = AutomatedTest(
    name="Unit Test",
    steps = [
        TestStep(
            name="Inject Flight 1",
            inject_flight=FlightInjectionAttempt(
                reference_time="2022-02-11T09:00:05.359502+00:00",
                name="f0001",
                test_injection=InjectFlightRequest(
                    operational_intent=None,
                    flight_authorisation=None
                ),
                known_responses=KnownResponses(
                    acceptable_results=[],
                    incorrect_result_details={}
                ),
                injection_target=InjectionTarget(uss_role="First-Mover USS")
            )
        ),
        TestStep(
            name="Inject Flight 2",
            inject_flight=FlightInjectionAttempt(
                reference_time="2022-02-11T09:30:05.359502+00:00",
                name="f0002",
                test_injection=InjectFlightRequest(
                    operational_intent=None,
                    flight_authorisation=None
                ),
                known_responses=KnownResponses(
                    acceptable_results=[],
                    incorrect_result_details={}
                ),
                injection_target=InjectionTarget(uss_role="Second USS")
            )
        ),
        TestStep(
            name="Delete Flight",
            delete_flight=FlightDeletionAttempt(
                flight_name="f0001"
            )
        )
    ]
)

def test_TestRunner():
    runner = TestRunner(auth_spec="NoAuth()", automated_test_id=automated_test.name, automated_test=automated_test, targets=targets)
    for s in automated_test.steps:
        print(s.name)
        runner.get_target(s)


def test_target_combinations():
    targets = [
        InjectionTargetConfiguration(
            name="uss_unit_test_1",
            injection_base_url="http://host.docker.internal:8075/scdsc"
        ),
        InjectionTargetConfiguration(
            name="uss_unit_test_2",
            injection_base_url="http://host.docker.internal:8076/scdsc"
        )
    ]

    targets_under_test = list(targets_combination(targets, automated_test.steps))
    assert len(targets_under_test) == 2
    assert targets_under_test[0]["First-Mover USS"].name == "uss_unit_test_1"
    assert targets_under_test[0]["Second USS"].name == "uss_unit_test_2"
    assert targets_under_test[1]["First-Mover USS"].name == "uss_unit_test_2"
    assert targets_under_test[1]["Second USS"].name == "uss_unit_test_1"

