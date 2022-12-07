import json
from typing import List
from implicitdict import ImplicitDict, StringBasedDateTime
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.resources.eurocae.ed269.source_document import (
    SourceDocument,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenario
from uas_standards.eurocae_ed269 import UASZoneVersion


# TODO: When the format is confirmed, this should be moved to uas_standards.eurocae_ed269
class ED269SchemaFile(ImplicitDict):
    formatVersion: str
    createdAt: StringBasedDateTime
    UASZoneList: List[UASZoneVersion]


class SourceDataModelValidation(TestScenario):
    source_document: SourceDocument

    def __init__(self, source_document: SourceDocument):
        super().__init__()
        self.source_document = source_document

    def run(self):
        self.begin_test_scenario()

        self.record_note(
            "Document",
            f"Ready at {self.source_document.specification.url}",
        )

        self.begin_test_case("ED-269 data model compliance")
        self.begin_test_step("Valid source")

        data=None
        with self.check("Valid JSON", [self.source_document.specification.url]) as check:
            try:
                data = json.loads(self.source_document.raw_document)
            except json.decoder.JSONDecodeError as e:
                check.record_failed(
                    summary="Unable to deserialize the document as JSON",
                    severity=Severity.High,
                    details=str(e),
                )

        if data:
            with self.check(
                "Valid schema and values", [self.source_document.specification.url]
            ) as check:
                try:
                    ImplicitDict.parse(data, ED269SchemaFile)
                except ValueError as e:
                    check.record_failed(
                        summary="Invalid format error",
                        severity=Severity.High,
                        details=str(e),
                    )

        self.end_test_step()
        self.end_test_case()
        self.end_test_scenario()
