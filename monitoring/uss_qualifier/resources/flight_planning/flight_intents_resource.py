from datetime import timedelta
import json
from typing import List

import arrow
from implicitdict import ImplicitDict, StringBasedDateTime

from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
)
from monitoring.uss_qualifier.fileio import load_dict
from monitoring.uss_qualifier.resources.resource import Resource
from monitoring.uss_qualifier.resources.flight_planning.flight_intent import (
    FlightIntentCollection,
    FlightIntentsSpecification,
)


class FlightIntentsResource(Resource[FlightIntentsSpecification]):
    _planning_time: timedelta
    _intent_collection: FlightIntentCollection

    def __init__(self, specification: FlightIntentsSpecification):
        self._intent_collection = ImplicitDict.parse(
            load_dict(specification.file_source), FlightIntentCollection
        )
        self._planning_time = specification.planning_time.timedelta

    def get_flight_intents(self) -> List[InjectFlightRequest]:
        t0 = arrow.utcnow() + self._planning_time

        intents: List[InjectFlightRequest] = []

        for intent in self._intent_collection.intents:
            dt = t0 - intent.reference_time.datetime
            req: InjectFlightRequest = ImplicitDict.parse(
                json.loads(json.dumps(intent.request)), InjectFlightRequest
            )

            for volume in (
                req.operational_intent.volumes
                + req.operational_intent.off_nominal_volumes
            ):
                volume.time_start.value = StringBasedDateTime(
                    volume.time_start.value.datetime + dt
                )
                volume.time_end.value = StringBasedDateTime(
                    volume.time_end.value.datetime + dt
                )
            intents.append(req)

        return intents
