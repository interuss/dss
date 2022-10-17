from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
    InjectFlightResult,
)
from monitoring.uss_qualifier.scd.configuration import InjectionTargetConfiguration
from monitoring.uss_qualifier.resources.flight_planning.automated_test import (
    AutomatedTest,
    TestStep,
    FlightInjectionAttempt,
    InjectionTarget,
    FlightDeletionAttempt,
    KnownResponses,
)
from monitoring.uss_qualifier.scd.executor.executor import combine_targets
from monitoring.uss_qualifier.scd.executor.target import TestTarget

# Constants
FirstMoverRole = "First-Mover USS"
SecondMoverRole = "Second-Mover USS"

# Test data definition
automated_test = AutomatedTest(
    name="Unit Test",
    steps=[
        TestStep(
            name="Inject Flight 1",
            inject_flight=FlightInjectionAttempt(
                reference_time="2022-02-11T09:00:05.359502+00:00",
                name="f0001",
                test_injection=InjectFlightRequest(
                    operational_intent=None, flight_authorisation=None
                ),
                known_responses=KnownResponses(
                    acceptable_results=[InjectFlightResult.Planned],
                    incorrect_result_details={},
                ),
                injection_target=InjectionTarget(uss_role=FirstMoverRole),
            ),
        ),
        TestStep(
            name="Inject Flight 2",
            inject_flight=FlightInjectionAttempt(
                reference_time="2022-02-11T09:30:05.359502+00:00",
                name="f0002",
                test_injection=InjectFlightRequest(
                    operational_intent=None, flight_authorisation=None
                ),
                known_responses=KnownResponses(
                    acceptable_results=[InjectFlightResult.Planned],
                    incorrect_result_details={},
                ),
                injection_target=InjectionTarget(uss_role=SecondMoverRole),
            ),
        ),
        TestStep(
            name="Delete Flight",
            delete_flight=FlightDeletionAttempt(flight_name="f0001"),
        ),
    ],
)

injection_targets = [
    InjectionTargetConfiguration(
        name="uss_unit_test_1",
        injection_base_url="http://host.docker.internal:8075/scdsc",
    ),
    InjectionTargetConfiguration(
        name="uss_unit_test_2",
        injection_base_url="http://host.docker.internal:8076/scdsc",
    ),
]
configured_targets = list(
    map(lambda t: TestTarget(t.name, t, "NoAuth()"), injection_targets)
)
# End of Test data definition


def test_target_combinations():
    targets_under_test = list(combine_targets(configured_targets, automated_test.steps))
    assert len(targets_under_test) == 2
    assert targets_under_test[0][FirstMoverRole].name == "uss_unit_test_1"
    assert targets_under_test[0][SecondMoverRole].name == "uss_unit_test_2"
    assert targets_under_test[1][FirstMoverRole].name == "uss_unit_test_2"
    assert targets_under_test[1][SecondMoverRole].name == "uss_unit_test_1"
