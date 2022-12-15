import time
import uuid
from dataclasses import dataclass
import datetime
from enum import Enum
from typing import List, Dict, Optional, Set

from implicitdict import StringBasedDateTime
from monitoring.monitorlib import fetch
from uas_standards.astm.f3411.v19.api import (
    LatLngPoint,
    CreateIdentificationServiceAreaParameters,
    Volume4D,
    Volume3D,
    GeoPolygon,
    CreateSubscriptionParameters,
    SubscriptionCallbacks,
    UpdateIdentificationServiceAreaParameters,
)

from monitoring.monitorlib.infrastructure import UTMClientSession
from monitoring.monitorlib.rid import SCOPE_READ, SCOPE_WRITE
from monitoring.monitorlib.rid_common import RIDVersion
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.reports.report import ParticipantID
from monitoring.uss_qualifier.resources.astm.f3411.dss import (
    DSSInstancesResource,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenario


VERTICES = [
    LatLngPoint(lng=130.6205, lat=-23.6558),
    LatLngPoint(lng=130.6301, lat=-23.6898),
    LatLngPoint(lng=130.6700, lat=-23.6709),
    LatLngPoint(lng=130.6466, lat=-23.6407),
]

GEO_POLYGON_STRING = ",".join("{},{}".format(x["lat"], x["lng"]) for x in VERTICES)

SHORT_WAIT_SEC = 5


class EntityType(str, Enum):
    ISA = "ISA"
    Sub = "Sub"


@dataclass
class TestEntity(object):
    type: EntityType
    uuid: str
    version: Optional[str] = None


TestObject = str


def _make_vol4(time_end: datetime.datetime) -> Volume4D:
    return Volume4D(
        spatial_volume=Volume3D(
            footprint=GeoPolygon(vertices=VERTICES),
            altitude_lo=20,
            altitude_hi=400,
        ),
        time_end=StringBasedDateTime(time_end),
    )


def _make_create_isa(
    time_end: datetime.datetime,
) -> CreateIdentificationServiceAreaParameters:
    return CreateIdentificationServiceAreaParameters(
        extents=_make_vol4(time_end),
        flights_url="https://example.com/uss/flights",
    )


def _make_create_subscription(
    time_end: datetime.datetime,
) -> CreateSubscriptionParameters:
    return CreateSubscriptionParameters(
        extents=_make_vol4(time_end),
        callbacks=SubscriptionCallbacks(
            identification_service_area_url="https://example.com/uss/identification_service_area",
        ),
    )


def _make_update_isa(
    time_end: datetime.datetime,
) -> UpdateIdentificationServiceAreaParameters:
    return UpdateIdentificationServiceAreaParameters(
        extents=_make_vol4(time_end),
        flights_url="https://example.com/uss/flights",
    )


def _extract_sub_ids_from_isa_put_response(response: dict) -> Set[str]:
    returned_subs = set()
    for subscriber in response["subscribers"]:
        for sub in subscriber["subscriptions"]:
            returned_subs.add(sub["subscription_id"])
    return returned_subs


class DSSInteroperability(TestScenario):
    _primary_dss_instance: ParticipantID
    _other_dss_instances: List[ParticipantID]
    _dss_map: Dict[ParticipantID, UTMClientSession]
    _context: Dict[TestObject, TestEntity]

    def __init__(
        self,
        dss_instances: DSSInstancesResource,
    ):
        super().__init__()
        # TODO: Add DSS combinations action generator to rotate primary DSS instance
        if any(
            dss.rid_version != RIDVersion.f3411_19
            for dss in dss_instances.dss_instances
        ):
            raise ValueError(
                f"DSSInteroperability scenario currently only supports DSS instances compliant with ASTM F3411-19"
            )
            # TODO: Expand support to F3411-22a
        self._primary_dss_instance = dss_instances.dss_instances[0].participant_id
        self._other_dss_instances = [
            dss.participant_id for dss in dss_instances.dss_instances[1:]
        ]
        self._dss_map = {
            dss.participant_id: dss.client for dss in dss_instances.dss_instances
        }
        self._context: Dict[TestObject, TestEntity] = {}

    def run(self):
        self.begin_test_scenario()

        if self._other_dss_instances:
            self.begin_test_case("Interoperability sequence")

            for i in range(1, 18):
                self.begin_test_step(f"S{i}")
                step = getattr(self, f"step{i}")
                step()
                self.end_test_step()

            self.end_test_case()

        self.end_test_scenario()

    def step1(self):
        """Create ISA in Primary DSS with 10 min TTL."""

        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
        isa1 = TestEntity(EntityType.ISA, str(uuid.uuid4()))
        self._context["isa_1"] = isa1
        initiated_at = datetime.datetime.utcnow()
        resp = self._dss_map[self._primary_dss_instance].put(
            f"/v1/dss/identification_service_areas/{isa1.uuid}",
            json=_make_create_isa(time_end),
            scope=SCOPE_WRITE,
        )
        query = fetch.describe_query(resp, initiated_at)
        self.record_query(query)
        with self.check(
            "ISA[P] created with proper response", [self._primary_dss_instance]
        ) as check:
            if resp.status_code != 200:
                check.record_failed(
                    f"Failed to insert ISA to {self._primary_dss_instance}",
                    Severity.High,
                    details=f"{resp.status_code} response: " + resp.content.decode(),
                    query_timestamps=[query.request.timestamp],
                )

        # save ISA_1 Version String
        isa1.version = resp.json()["service_area"]["version"]

    def step2(self):
        """Can create Subscription in all DSSs, ISA accessible from all
        non-primary DSSs."""
        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
        for index, dss in enumerate(
            [self._primary_dss_instance] + self._other_dss_instances
        ):
            sub_1_uuid = str(uuid.uuid4())
            sub1 = TestEntity(EntityType.Sub, sub_1_uuid)
            self._context[f"sub_1_{index}"] = sub1
            initiated_at = datetime.datetime.utcnow()
            resp = self._dss_map[dss].put(
                f"/v1/dss/subscriptions/{sub_1_uuid}",
                json=_make_create_subscription(time_end),
                scope=SCOPE_READ,
            )
            query = fetch.describe_query(resp, initiated_at)
            self.record_query(query)
            with self.check(
                "Subscription[n] created with proper response", [dss]
            ) as check:
                if resp.status_code != 200:
                    check.record_failed(
                        f"Failed to Insert Subscription to {dss}: {resp.content.decode()}",
                        Severity.High,
                        details=f"{resp.status_code} response: "
                        + resp.content.decode(),
                        query_timestamps=[query.request.timestamp],
                    )

            data = resp.json()
            isa_ids = [isa["id"] for isa in data["service_areas"]]
            isa1_uuid = self._context[f"isa_1"].uuid
            with self.check("service_areas includes ISA from S1", [dss]) as check:
                if isa1_uuid not in isa_ids:
                    check.record_failed(
                        f"{dss} did not return ISA from testStep1 when creating Subscription",
                        Severity.High,
                        details="service_areas IDs: "
                        + ", ".join(isa_ids)
                        + f"; isa_1 ID: {isa1_uuid}",
                        query_timestamps=[query.request.timestamp],
                    )

            # save SUB_1 Version String
            sub1.version = data["subscription"]["version"]

    def step3(self):
        """Can retrieve specific Subscription emplaced in primary DSS
        from all DSSs."""
        for dss in [self._primary_dss_instance] + self._other_dss_instances:
            initiated_at = datetime.datetime.utcnow()
            resp = self._dss_map[dss].get(
                f"/v1/dss/subscriptions/{self._context['sub_1_0'].uuid}",
                scope=SCOPE_READ,
            )
            query = fetch.describe_query(resp, initiated_at)
            self.record_query(query)
            with self.check(
                "Subscription[P] returned with proper response", [dss]
            ) as check:
                if resp.status_code != 200:
                    check.record_failed(
                        f"{dss} failed to get Subscription emplaced in {self._primary_dss_instance}",
                        Severity.High,
                        details=f"{resp.status_code} response: "
                        + resp.content.decode(),
                        query_timestamps=[query.request.timestamp],
                    )

                data = resp.json()

                if self._context["sub_1_0"].uuid != data["subscription"]["id"]:
                    check.record_failed(
                        f"{dss} did not return correct Subscription",
                        Severity.High,
                        details=f"Expected Subscription ID {self._context['sub_1_0'].uuid} but got {data['subscription']['id']}",
                        query_timestamps=[query.request.timestamp],
                    )

    def step4(self):
        """Can query all Subscriptions in area from all DSSs."""
        all_dss = [self._primary_dss_instance] + self._other_dss_instances
        all_sub_1 = set()
        for index in range(len(all_dss)):
            all_sub_1.add(self._context[f"sub_1_{index}"].uuid)
        for index, dss in enumerate(all_dss):
            initiated_at = datetime.datetime.utcnow()
            resp = self._dss_map[dss].get(
                f"/v1/dss/subscriptions?area={GEO_POLYGON_STRING}", scope=SCOPE_READ
            )
            query = fetch.describe_query(resp, initiated_at)
            self.record_query(query)
            with self.check(
                "Can query all Subscriptions in area from all DSSs", [dss]
            ) as check:
                if resp.status_code != 200:
                    check.record_failed(
                        f"{dss} failed to get Subscription 1 by area",
                        Severity.High,
                        details=f"{resp.status_code} response: "
                        + resp.content.decode(),
                        query_timestamps=[query.request.timestamp],
                    )

                returned_subs = set([x["id"] for x in resp.json()["subscriptions"]])
                missing_subs = all_sub_1 - returned_subs
                if missing_subs:
                    check.record_failed(
                        f"{dss} returned too few Subscriptions",
                        Severity.High,
                        details=f"Missing: {', '.join(missing_subs)}",
                        query_timestamps=[query.request.timestamp],
                    )

    def step5(self):
        """Can modify ISA in primary DSS, ISA modification triggers
        subscription notification requests"""
        isa_1 = self._context["isa_1"]
        initiated_at = datetime.datetime.utcnow()
        resp = self._dss_map[self._primary_dss_instance].get(
            f"/v1/dss/identification_service_areas/{isa_1.uuid}", scope=SCOPE_READ
        )
        query = fetch.describe_query(resp, initiated_at)
        self.record_query(query)
        with self.check(
            "Can get ISA from primary DSS", [self._primary_dss_instance]
        ) as check:
            if resp.status_code != 200:
                check.record_failed(
                    f"Failed to find ISA_1 in Primary DSS {self._primary_dss_instance}",
                    Severity.High,
                    details=f"{resp.status_code} response: " + resp.content.decode(),
                    query_timestamps=[query.request.timestamp],
                )
        data = resp.json()["service_area"]
        isa_1.version = data["version"]

        time_end = datetime.datetime.utcnow() + datetime.timedelta(
            seconds=SHORT_WAIT_SEC
        )
        isa_update = _make_update_isa(time_end)
        initiated_at = datetime.datetime.utcnow()
        put_resp = self._dss_map[self._primary_dss_instance].put(
            f"/v1/dss/identification_service_areas/{isa_1.uuid}/{isa_1.version}",
            json=isa_update,
            scope=SCOPE_WRITE,
        )
        query = fetch.describe_query(put_resp, initiated_at)
        self.record_query(query)
        with self.check(
            "Can modify ISA in primary DSS", [self._primary_dss_instance]
        ) as check:
            if put_resp.status_code != 200:
                check.record_failed(
                    f"Failed to update ISA_1 with new end time",
                    Severity.High,
                    details=f"{resp.status_code} response: " + resp.content.decode(),
                    query_timestamps=[query.request.timestamp],
                )

            updated_data = put_resp.json()
            t_updated = StringBasedDateTime(updated_data["service_area"]["time_end"])
            if t_updated.datetime != isa_update.extents.time_end.datetime:
                check.record_failed(
                    f"Unsuccessful Update; end time not changed as requested",
                    Severity.High,
                    f"Expected: '{isa_update.extents.time_end}', received: '{updated_data['service_area']['time_end']}'",
                )

        # TODO: Implement "ISA modification triggers subscription notification requests check"

    def step6(self):
        """Can delete all Subscription in primary DSS"""
        dss = self._dss_map[self._primary_dss_instance]
        for name, entity in self._context.items():
            if name.startswith("sub_1_"):
                initiated_at = datetime.datetime.utcnow()
                resp = dss.delete(
                    f"/v1/dss/subscriptions/{entity.uuid}/{entity.version}",
                    scope=SCOPE_READ,
                )
                query = fetch.describe_query(resp, initiated_at)
                self.record_query(query)
                with self.check(
                    "Subscription[n] deleted with proper response",
                    [self._primary_dss_instance],
                ) as check:
                    if resp.status_code != 200:
                        check.record_failed(
                            f"Failed to delete {entity.uuid} from Primary DSS {self._primary_dss_instance}",
                            Severity.High,
                            details=f"{resp.status_code} response: "
                            + resp.content.decode(),
                            query_timestamps=[query.request.timestamp],
                        )

    def step7(self):
        """Subscription deletion from ID index was effective on primary DSS"""
        all_dss = [self._primary_dss_instance] + self._other_dss_instances
        for index, dss_name in enumerate(all_dss):
            dss = self._dss_map[dss_name]
            sub_uuid = self._context[f"sub_1_{index}"].uuid
            initiated_at = datetime.datetime.utcnow()
            resp = dss.get(f"/v1/dss/subscriptions/{sub_uuid}", scope=SCOPE_READ)
            query = fetch.describe_query(resp, initiated_at)
            self.record_query(query)
            with self.check("404 with proper response", [dss_name]) as check:
                if resp.status_code != 404:
                    check.record_failed(
                        f"Expecting code 404, found {resp.status_code}",
                        Severity.High,
                        query_timestamps=[query.request.timestamp],
                    )

    def step8(self):
        """Subscription deletion from geographic index was effective on primary DSS"""
        all_sub_1 = [sub for sub in self._context if sub.startswith("sub_1_")]
        all_dss = [self._primary_dss_instance] + self._other_dss_instances
        for dss in all_dss:
            initiated_at = datetime.datetime.utcnow()
            resp = self._dss_map[dss].get(
                f"/v1/dss/subscriptions?area={GEO_POLYGON_STRING}", scope=SCOPE_READ
            )
            query = fetch.describe_query(resp, initiated_at)
            self.record_query(query)
            with self.check("Subscriptions queried successfully", [dss]) as check:
                if resp.status_code != 200:
                    check.record_failed(
                        f"Expecting code 200, found {resp.status_code}",
                        Severity.High,
                        f"Content: {resp.content.decode()}",
                        query_timestamps=[query.request.timestamp],
                    )

            data = resp.json()
            found_deleted_sub = [
                sub["id"] for sub in data["subscriptions"] if sub["id"] in all_sub_1
            ]

            with self.check(
                "No Subscription[i] 1≤i≤n returned with proper response", [dss]
            ) as check:
                if found_deleted_sub:
                    check.record_failed(
                        "Found deleted Subscriptions",
                        Severity.High,
                        f"Deleted Subscriptions found: {found_deleted_sub}",
                        query_timestamps=[query.request.timestamp],
                    )

    def step9(self):
        """Expired ISA automatically removed, ISA modifications
        accessible from all non-primary DSSs"""

        # sleep X seconds for ISA_1 to expire
        time.sleep(SHORT_WAIT_SEC)

        time_end = datetime.datetime.utcnow() + datetime.timedelta(
            seconds=SHORT_WAIT_SEC
        )
        for index, dss in enumerate(
            [self._primary_dss_instance] + self._other_dss_instances
        ):
            sub_2_uuid = str(uuid.uuid4())
            self._context[f"sub_2_{index}"] = TestEntity(EntityType.Sub, sub_2_uuid)
            initiated_at = datetime.datetime.utcnow()
            resp = self._dss_map[dss].put(
                f"/v1/dss/subscriptions/{sub_2_uuid}",
                json=_make_create_subscription(time_end),
                scope=SCOPE_READ,
            )
            query = fetch.describe_query(resp, initiated_at)
            self.record_query(query)
            with self.check(
                "Subscription[n] created with proper response", [dss]
            ) as check:
                if resp.status_code != 200:
                    check.record_failed(
                        f"Failed to Insert Subscription to {dss}",
                        Severity.High,
                        f"Content: {resp.content.decode()}",
                        query_timestamps=[query.request.timestamp],
                    )
            data = resp.json()
            isa_ids = [isa["id"] for isa in data["service_areas"]]
            with self.check(
                "service_areas does not include ISA from S1", [dss]
            ) as check:
                if self._context["isa_1"].uuid not in isa_ids:
                    pass  # TODO: Enforce check regarding return of expired ISAs
                    # check.record_failed(
                    #     f"{dss} returned expired ISA_1 when creating Subscription",
                    #     Severity.High,
                    # )

            # save SUB_2 Version String
            self._context[f"sub_2_{index}"].version = data["subscription"]["version"]

    def step10(self):
        """ISA creation triggers subscription notification requests"""

        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
        self._context["isa_2"] = TestEntity(EntityType.ISA, str(uuid.uuid4()))
        initiated_at = datetime.datetime.utcnow()
        resp = self._dss_map[self._primary_dss_instance].put(
            f"/v1/dss/identification_service_areas/{self._context['isa_2'].uuid}",
            json=_make_create_isa(time_end),
            scope=SCOPE_WRITE,
        )
        query = fetch.describe_query(resp, initiated_at)
        self.record_query(query)
        with self.check(
            "ISA[P] created with proper response", [self._primary_dss_instance]
        ) as check:
            if resp.status_code != 200:
                check.record_failed(
                    f"Failed to insert ISA to {self._primary_dss_instance}",
                    Severity.High,
                    f"Content: {resp.content.decode()}",
                    query_timestamps=[query.request.timestamp],
                )

        all_sub_2 = set()
        for name, entity in self._context.items():
            if name.startswith("sub_2_"):
                all_sub_2.add(entity.uuid)

        returned_subs = _extract_sub_ids_from_isa_put_response(resp.json())
        missing_subs = all_sub_2 - returned_subs
        with self.check(
            "All Subscription[i] 1≤i≤n returned in subscribers",
            [self._primary_dss_instance],
        ) as check:
            if missing_subs:
                check.record_failed(
                    f"{self._primary_dss_instance} returned too few Subscriptions",
                    Severity.High,
                    f"Missing Subscriptions: {', '.join(missing_subs)}",
                    query_timestamps=[query.request.timestamp],
                )

        # save ISA_2 Version String
        self._context["isa_2"].version = resp.json()["service_area"]["version"]

    def step11(self):
        """ISA deletion triggers subscription notification requests"""
        isa_2_uuid = self._context["isa_2"].uuid
        version = self._context["isa_2"].version
        initiated_at = datetime.datetime.utcnow()
        resp = self._dss_map[self._primary_dss_instance].delete(
            f"/v1/dss/identification_service_areas/{isa_2_uuid}/{version}",
            scope=SCOPE_WRITE,
        )
        query = fetch.describe_query(resp, initiated_at)
        self.record_query(query)
        with self.check(
            "ISA[P] deleted with proper response", [self._primary_dss_instance]
        ) as check:
            if resp.status_code != 200:
                check.record_failed(
                    f"Failed to delete ISA to {self._primary_dss_instance}",
                    Severity.High,
                    f"Content: {resp.content.decode()}",
                    query_timestamps=[query.request.timestamp],
                )

        all_sub_2 = set()
        for name, entity in self._context.items():
            if name.startswith("sub_2_"):
                all_sub_2.add(entity.uuid)

        returned_subs = _extract_sub_ids_from_isa_put_response(resp.json())
        missing_subs = all_sub_2 - returned_subs
        with self.check(
            "All Subscription[i] 1≤i≤n returned in subscribers",
            [self._primary_dss_instance],
        ) as check:
            if missing_subs:
                check.record_failed(
                    f"{self._primary_dss_instance} returned too few Subscriptions",
                    Severity.High,
                    f"Missing Subscriptions: {', '.join(missing_subs)}",
                    query_timestamps=[query.request.timestamp],
                )

    def step12(self):
        """Expired Subscriptions don’t trigger subscription notification requests"""
        time.sleep(SHORT_WAIT_SEC)
        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
        self._context["isa_3"] = TestEntity(EntityType.ISA, str(uuid.uuid4()))
        initiated_at = datetime.datetime.utcnow()
        resp = self._dss_map[self._primary_dss_instance].put(
            f"/v1/dss/identification_service_areas/{self._context['isa_3'].uuid}",
            json=_make_create_isa(time_end),
            scope=SCOPE_WRITE,
        )
        query = fetch.describe_query(resp, initiated_at)
        self.record_query(query)
        with self.check(
            "ISA[P] created with proper response", [self._primary_dss_instance]
        ) as check:
            if resp.status_code != 200:
                check.record_failed(
                    f"Failed to insert ISA to {self._primary_dss_instance}",
                    Severity.High,
                    f"Content: {resp.content.decode()}",
                    query_timestamps=[query.request.timestamp],
                )

        all_sub_2 = [sub for sub in self._context if sub.startswith("sub_2_")]
        returned_subs = _extract_sub_ids_from_isa_put_response(resp.json())
        found_expired_sub = [sub for sub in returned_subs if sub in all_sub_2]
        with self.check(
            "None of Subscription[i] 1≤i≤n returned in subscribers",
            [self._primary_dss_instance],
        ) as check:
            if found_expired_sub:
                check.record_failed(
                    "Found expired Subscriptions",
                    Severity.High,
                    f"Expired Subscriptions found: {', '.join(found_expired_sub)}",
                    query_timestamps=[query.request.timestamp],
                )

        # save ISA_3 Version String
        self._context["isa_3"].version = resp.json()["service_area"]["version"]

    def step13(self):
        """Expired Subscription removed from geographic index on primary DSS"""
        all_dss = [self._primary_dss_instance] + self._other_dss_instances
        all_sub_2 = set()
        for index in range(len(all_dss)):
            all_sub_2.add(self._context[f"sub_2_{index}"].uuid)
        for index, dss in enumerate(all_dss):
            initiated_at = datetime.datetime.utcnow()
            resp = self._dss_map[dss].get(
                f"/v1/dss/subscriptions?area={GEO_POLYGON_STRING}", scope=SCOPE_READ
            )
            query = fetch.describe_query(resp, initiated_at)
            self.record_query(query)
            assert (
                resp.status_code == 200
            ), f"{dss} failed to get SUB_2 by area: {resp.content.decode()}"

            returned_subs = set([x["id"] for x in resp.json()["subscriptions"]])
            found_expired_sub = [sub for sub in returned_subs if sub in all_sub_2]

            with self.check(
                "No Subscription[i] 1≤i≤n returned with proper response", [dss]
            ) as check:
                if found_expired_sub:
                    check.record_failed(
                        "Found expired Subscriptions",
                        Severity.High,
                        f"Expired Subscriptions found: {', '.join(found_expired_sub)}",
                        query_timestamps=[query.request.timestamp],
                    )

    def step14(self):
        """Expired Subscription still accessible shortly after expiration"""
        all_dss = [self._primary_dss_instance] + self._other_dss_instances
        for index, dss in enumerate(all_dss):
            sub_2_uuid = self._context[f"sub_2_{index}"].uuid
            initiated_at = datetime.datetime.utcnow()
            resp = self._dss_map[dss].get(
                f"/v1/dss/subscriptions/{sub_2_uuid}", scope=SCOPE_READ
            )
            query = fetch.describe_query(resp, initiated_at)
            self.record_query(query)
            # TODO: Investigate expected behavior and "404 with proper response" check

    def step15(self):
        """ISA deletion does not trigger subscription
        notification requests for expired Subscriptions"""
        isa_3_uuid = self._context["isa_3"].uuid
        version = self._context["isa_3"].version
        initiated_at = datetime.datetime.utcnow()
        resp = self._dss_map[self._primary_dss_instance].delete(
            f"/v1/dss/identification_service_areas/{isa_3_uuid}/{version}",
            scope=SCOPE_WRITE,
        )
        query = fetch.describe_query(resp, initiated_at)
        self.record_query(query)
        with self.check(
            "ISA[P] deleted with proper response", [self._primary_dss_instance]
        ) as check:
            if resp.status_code != 200:
                check.record_failed(
                    f"Failed to delete ISA_3 from {self._primary_dss_instance}",
                    Severity.High,
                    f"{resp.status_code} response: {resp.content.decode()}",
                    query_timestamps=[query.request.timestamp],
                )

        all_sub_2 = [sub for sub in self._context if sub.startswith("sub_2_")]
        returned_subs = _extract_sub_ids_from_isa_put_response(resp.json())
        found_expired_sub = [sub for sub in returned_subs if sub in all_sub_2]
        with self.check(
            "None of Subscription[i] 1≤i≤n returned in subscribers with proper response",
            [self._primary_dss_instance],
        ) as check:
            if found_expired_sub:
                check.record_failed(
                    "Found expired Subscriptions",
                    Severity.High,
                    f"Expired Subscriptions found: {', '.join(found_expired_sub)}",
                    query_timestamps=[query.request.timestamp],
                )

    def step16(self):
        """Deleted ISA removed from all DSSs"""
        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
        for index, dss in enumerate(
            [self._primary_dss_instance] + self._other_dss_instances
        ):
            sub_3_uuid = str(uuid.uuid4())
            self._context[f"sub_3_{index}"] = TestEntity(EntityType.Sub, sub_3_uuid)
            initiated_at = datetime.datetime.utcnow()
            resp = self._dss_map[dss].put(
                f"/v1/dss/subscriptions/{sub_3_uuid}",
                json=_make_create_subscription(time_end),
                scope=SCOPE_READ,
            )
            query = fetch.describe_query(resp, initiated_at)
            self.record_query(query)
            with self.check(
                "Subscription[n] created with proper response", [dss]
            ) as check:
                if resp.status_code != 200:
                    check.record_failed(
                        f"Failed to Insert Subscription to {dss}",
                        Severity.High,
                        f"Content: {resp.content.decode()}",
                        query_timestamps=[query.request.timestamp],
                    )
            data = resp.json()
            isa_ids = [isa["id"] for isa in data["service_areas"]]
            with self.check(
                "service_areas does not include ISA from S12", [dss]
            ) as check:
                if self._context["isa_3"].uuid in isa_ids:
                    check.record_failed(
                        f"{dss} returned deleted ISA_3 when creating Subscription",
                        Severity.Medium,
                        query_timestamps=[query.request.timestamp],
                    )

            # save SUB_3 Version String
            self._context[f"sub_3_{index}"].version = data["subscription"]["version"]

    def step17(self):
        """Clean up SUBS_3"""
        for name, sub_3 in self._context.items():
            if not name.startswith("sub_3_"):
                continue
            initiated_at = datetime.datetime.utcnow()
            resp = self._dss_map[self._primary_dss_instance].delete(
                f"/v1/dss/subscriptions/{sub_3.uuid}/{sub_3.version}", scope=SCOPE_READ
            )
            query = fetch.describe_query(resp, initiated_at)
            self.record_query(query)
            with self.check(
                "Subscription[n] deleted with proper response",
                [self._primary_dss_instance],
            ) as check:
                if resp.status_code != 200:
                    check.record_failed(
                        f"Failed to delete SUB_3 from Primary DSS {self._primary_dss_instance}",
                        Severity.High,
                        f"{resp.status_code} response: {resp.content.decode()}",
                        query_timestamps=[query.request.timestamp],
                    )

    def cleanup(self):
        self.begin_cleanup()

        dss = self._dss_map[self._primary_dss_instance]
        for entity in self._context.values():
            if entity.type == EntityType.ISA:
                initiated_at = datetime.datetime.utcnow()
                resp = dss.get(
                    f"/v1/dss/identification_service_areas/{entity.uuid}",
                    scope=SCOPE_READ,
                )
                query = fetch.describe_query(resp, initiated_at)
                self.record_query(query)
                if resp.status_code == 404:
                    continue
                resp.raise_for_status()
                entity.version = resp.json()["service_area"]["version"]

                initiated_at = datetime.datetime.utcnow()
                resp = dss.delete(
                    f"/v1/dss/identification_service_areas/{entity.uuid}/{entity.version}",
                    scope=SCOPE_WRITE,
                )
                query = fetch.describe_query(resp, initiated_at)
                self.record_query(query)
                with self.check(
                    "ISA deleted with proper response",
                    [self._primary_dss_instance],
                ) as check:
                    if resp.status_code != 200:
                        check.record_failed(
                            f"Could not clean up ISA {entity.uuid}/{entity.version}",
                            Severity.Medium,
                            f"{resp.status_code} response: {resp.content.decode()}",
                            query_timestamps=[query.request.timestamp],
                        )
            elif entity.type == EntityType.Sub:
                initiated_at = datetime.datetime.utcnow()
                resp = dss.get(
                    f"/v1/dss/subscriptions/{entity.uuid}",
                    scope=SCOPE_READ,
                )
                query = fetch.describe_query(resp, initiated_at)
                self.record_query(query)
                if resp.status_code == 404:
                    continue
                resp.raise_for_status()
                entity.version = resp.json()["subscription"]["version"]

                initiated_at = datetime.datetime.utcnow()
                resp = dss.delete(
                    f"/v1/dss/subscriptions/{entity.uuid}/{entity.version}",
                    scope=SCOPE_READ,
                )
                query = fetch.describe_query(resp, initiated_at)
                self.record_query(query)
                with self.check(
                    "Subscription deleted with proper response",
                    [self._primary_dss_instance],
                ) as check:
                    if resp.status_code != 200:
                        check.record_failed(
                            f"Could not clean up Subscription {entity.uuid}/{entity.version}",
                            Severity.Medium,
                            f"{resp.status_code} response: {resp.content.decode()}",
                            query_timestamps=[query.request.timestamp],
                        )
            else:
                raise RuntimeError(f"Unknown Entity type: {entity.type}")

        self.end_cleanup()
