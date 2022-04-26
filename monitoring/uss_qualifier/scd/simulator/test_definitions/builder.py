import json
import os
from typing import List

from monitoring.monitorlib.locality import Locality
from monitoring.uss_qualifier.scd.data_interfaces import RequiredUSSCapabilities, AutomatedTest, TestStep


TEST_DEFINITIONS_BASEDIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../', 'test_definitions'))

class AutomatedTestBuilder():
    name: str
    group: str
    locale: Locality
    uss_capabilities: RequiredUSSCapabilities
    steps: List[TestStep]

    def __init__(self, name: str, group: str, locale: Locality):
        self.name = name
        self.group = group
        self.locale = locale

    def build(self) -> AutomatedTest:
        return AutomatedTest(
            name=self.name,
            uss_capabilities=self.uss_capabilities,
            steps=self.steps
        )

    def save(self):
        output = self.build()
        output_path = self.get_output_path()
        print(f"Saving test '{self.name}' ({self.group}) to {output_path}")
        with open(output_path, 'w') as f:
            f.write(json.dumps(output))

    def get_filename(self):
        return self.name.replace(' ', '-').lower() + '.json'

    def get_output_path(self):
        return os.path.join(TEST_DEFINITIONS_BASEDIR, self.locale.upper(), self.group, self.get_filename())
