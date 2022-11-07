from typing import List

from monitoring.uss_qualifier.resources.astm.f3411.dss import (
    DSSInstancesResource,
    DSSInstance,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenario


class DSSInteroperability(TestScenario):
    _primary_dss_instance: DSSInstance
    _other_dss_instances: List[DSSInstance]

    def __init__(
        self,
        dss_instances: DSSInstancesResource,
    ):
        super().__init__()
        raise NotImplementedError()

    def run(self):
        raise NotImplementedError("TODO: adapt from monitoring/interoperability")
