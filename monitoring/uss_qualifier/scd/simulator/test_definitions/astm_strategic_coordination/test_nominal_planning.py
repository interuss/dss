import json

from monitoring.monitorlib.locality import Locality
from monitoring.uss_qualifier.scd.simulator.test_definitions.astm_strategic_coordination.nominal_planning import \
    NominalPlanningTestDefinition
from monitoring.uss_qualifier.scd.simulator.test_definitions.builder import TEST_DEFINITIONS_BASEDIR


def test_nominal_planning_output():
    with open(f"{TEST_DEFINITIONS_BASEDIR}/../test_definitions/CHE/astm-strategic-coordination/nominal-planning-1.json") as f:
        reference_data = json.load(f)
    test_definition = NominalPlanningTestDefinition(locale=Locality('CHE'))
    assert json.dumps(reference_data, sort_keys=True) == json.dumps(test_definition.build(), sort_keys=True)

