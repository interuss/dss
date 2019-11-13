import clients
import itertools
import inspect
import datetime
import uuid
import collections
import time
import logging

from typing import Dict, Any, List, Callable, Iterable

logging.basicConfig(level=logging.INFO)
LOG = logging.getLogger(__name__)

DATE_FORMAT = "%Y-%m-%dT%H:%M:%S.%fZ"

VERTICES = [
    {"lng": 130.6205, "lat": -23.6558},
    {"lng": 130.6301, "lat": -23.6898},
    {"lng": 130.6700, "lat": -23.6709},
    {"lng": 130.6466, "lat": -23.6407},
]

GEO_POLYGON_STRING = ",".join("{},{}".format(x["lat"], x["lng"]) for x in VERTICES)

SHORT_WAIT_SEC = 5

# "type" indicates the type of entity could be ISAs or SUBs
# "uuid" is the actual UUID value of entity stored in the DSS
TestContext = collections.namedtuple("TestContext", ["type", "uuid"])


class InterOpTestSuite:
    def __init__(self, dss_clients: Dict[str, clients.DSSClient]):
        self.dss_clients = dss_clients

    def startTest(self):
        for round, dss_permutation in enumerate(
            itertools.permutations(self.dss_clients)
        ):
            primary_dss = dss_permutation[0]
            all_other_dss = list(dss_permutation[1:])
            ts = TestSteps()
            LOG.info(f"Round {round}")
            for name, test_step in self._getTests().items():
                try:
                    test_step(
                        ts, self.dss_clients, primary_dss, all_other_dss=all_other_dss
                    )
                    LOG.info(f"{name} Passed with {primary_dss} as primary DSS")
                except AssertionError as e:
                    docstring = inspect.cleandoc(inspect.getdoc(test_step))
                    msg = (
                        f"Failed {name} with {primary_dss} as primary DSS\n"
                        f"\tTest Purpose: {docstring}\n"
                        f"\tFailure Message: {e}\n"
                        f"\tContinuing to next round if any"
                    )
                    LOG.error(msg)
                    LOG.debug(f"Cleaning up round {round + 1}")
                    ts.cleanUp(self.dss_clients, primary_dss)
                    break

            LOG.debug(f"Cleaning up round {round + 1}")
            ts.cleanUp(self.dss_clients, primary_dss)

    def _getTests(self) -> Dict[str, Callable]:
        # methods is a list of Tuples
        methods = inspect.getmembers(TestSteps, predicate=inspect.isfunction)
        result = {name: obj for name, obj in methods if name.startswith("testStep")}
        # sort by name
        ordered_result = collections.OrderedDict()
        for i in range(1, len(result) + 1):
            name = f"testStep{i}"
            ordered_result[name] = result[name]

        return ordered_result


class TestSteps:
    """Class containing Test Steps & Context for executing the tests
    functions that follows testStep%d naming convention will be run as test steps
    """

    def __init__(self):
        self.context: Dict[Any, TestContext] = {}

    def _extract_sub_ids_from_isa_put_response(
        self, response: Dict[str, Any]
    ) -> Iterable[str]:
        returned_subs = set()
        for subscriber in response["subscribers"]:
            for subscription in subscriber:
                for sub in subscriber["subscriptions"]:
                    returned_subs.add(sub["subscription_id"])
        return returned_subs

    def cleanUp(self, dss_map, primary_dss):
        dss = dss_map[primary_dss]
        for entity_type, stored_uuid in self.context.values():
            if entity_type == "ISA":
                version = self.context[stored_uuid].uuid
                dss.delete(f"/identification_service_areas/{stored_uuid}/{version}")
            elif entity_type == "SUB":
                version = self.context[stored_uuid].uuid
                dss.delete(f"/subscriptions/{stored_uuid}/{version}")
            elif entity_type == "VERSION":
                # do nothing
                continue
            else:
                LOG.warning(f"Unknown Type: {entity_type}")

    def testStep1(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """Create ISA in Primary DSS with 10 min TTL."""

        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
        self.context["isa_1_uuid"] = TestContext("ISA", str(uuid.uuid4()))
        resp = dss_map[primary_dss].put(
            f"/identification_service_areas/{self.context['isa_1_uuid'].uuid}",
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {"vertices": VERTICES},
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_end": time_end.strftime(DATE_FORMAT),
                },
                "flights_url": "https://example.com/uss/flights",
            },
        )
        assert (
            resp.status_code == 200
        ), f"Failed to insert ISA to {primary_dss}. Error: {resp.json()['error']}"

        # save ISA_1 Version String
        self.context[self.context["isa_1_uuid"].uuid] = TestContext(
            "VERSION",
            resp.json()["service_area"]["version"],
        )

    def testStep2(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Can create Subscription in all DSSs, ISA accessible from all
        non-primary DSSs."""
        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
        for index, dss in enumerate([primary_dss] + all_other_dss):
            sub_1_uuid = str(uuid.uuid4())
            self.context[f"sub_1_{index}_uuid"] = TestContext("SUB", sub_1_uuid)
            resp = dss_map[dss].put(
                f"/subscriptions/{sub_1_uuid}",
                json={
                    "extents": {
                        "spatial_volume": {
                            "footprint": {"vertices": VERTICES},
                            "altitude_lo": 20,
                            "altitude_hi": 400,
                        },
                        "time_end": time_end.strftime(DATE_FORMAT),
                    },
                    "callbacks": {
                        "identification_service_area_url": "https://example.com/uss/identification_service_area"
                    },
                },
            )
            assert resp.status_code == 200, f"Failed to Insert Subscription to {dss}"
            data = resp.json()
            isa_ids = [isa["id"] for isa in data["service_areas"]]
            assert (
                self.context["isa_1_uuid"].uuid in isa_ids
            ), f"{dss} did not return ISA from testStep1 when creating Subscription"

            # save SUB_1 Version String
            self.context[sub_1_uuid] = TestContext(
                "VERSION",
                data["subscription"]["version"],
            )

    def testStep3(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Can retrieve specific Subscription emplaced in primary DSS
        from all DSSs."""
        for dss in [primary_dss] + all_other_dss:
            resp = dss_map[dss].get(
                f"/subscriptions/{self.context['sub_1_0_uuid'].uuid}"
            )
            assert resp.status_code == 200, f"{dss} failed to get SUB_1"

            data = resp.json()
            assert (
                self.context["sub_1_0_uuid"].uuid == data["subscription"]["id"]
            ), f"{dss} did not return correct Subscription"

    def testStep4(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Can query all Subscriptions in area from all DSSs."""
        all_dss = [primary_dss] + all_other_dss
        all_sub_1 = set()
        for index in range(len(all_dss)):
            all_sub_1.add(self.context[f"sub_1_{index}_uuid"].uuid)
        for index, dss in enumerate(all_dss):
            resp = dss_map[dss].get(f"/subscriptions?area={GEO_POLYGON_STRING}")
            assert resp.status_code == 200, f"{dss} failed to get SUB_1 by area"

            returned_subs = set([x["id"] for x in resp.json()["subscriptions"]])

            missing_subs = all_sub_1 - returned_subs
            assert (
                missing_subs == set()
            ), f"{dss} returned too few Subscriptions, missing: {missing_subs}"

    def testStep5(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """Can modify ISA in primary DSS, ISA modification triggers 
        subscription notification requests"""
        isa_1_uuid = self.context["isa_1_uuid"].uuid
        resp = dss_map[primary_dss].get(f"/identification_service_areas/{isa_1_uuid}")
        assert (
            resp.status_code == 200
        ), f"Failed to find ISA_1 in Primary DSS: {primary_dss}"
        data = resp.json()["service_area"]

        time_end = datetime.datetime.utcnow() + datetime.timedelta(
            seconds=SHORT_WAIT_SEC
        )
        put_resp = dss_map[primary_dss].put(
            f"/identification_service_areas/{isa_1_uuid}/{data['version']}",
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {"vertices": VERTICES},
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_end": time_end.strftime(DATE_FORMAT),
                },
                "flights_url": "https://example.com/uss/flights",
            },
        )
        assert put_resp.status_code == 200, f"Failed to update ISA_1 with new end time"

        updated_data = put_resp.json()
        assert updated_data["service_area"]["time_end"] == time_end.strftime(
            DATE_FORMAT
        ), f"Unsuccessful Update; no change to end time"

    def testStep6(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """Can delete all Subscription in primary DSS"""
        dss = dss_map[primary_dss]
        for name, entity in self.context.items():
            if name.startswith("sub_1_"):
                version = self.context[entity.uuid].uuid
                resp = dss.delete(f"/subscriptions/{entity.uuid}/{version}")
                assert (
                    resp.status_code == 200
                ), f"Failed to delete {entity.uuid} from Primary DSS: {primary_dss}"

    def testStep7(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Subscription deletion from ID index was effective on primary DSS"""
        all_dss = [primary_dss] + all_other_dss
        for index, dss_name in enumerate(all_dss):
            dss = dss_map[dss_name]
            sub_uuid = self.context[f"sub_1_{index}_uuid"].uuid
            resp = dss.get(f"/subscriptions/{sub_uuid}")
            assert (
                resp.status_code == 404
            ), f"Expecting code 404, found {resp.status_code}"

    def testStep8(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Subscription deletion from geographic index was effective on primary DSS"""
        all_sub_1 = [sub for sub in self.context if sub.startswith("sub_1_")]
        all_dss = [primary_dss] + all_other_dss
        for dss in all_dss:
            resp = dss_map[dss].get(f"/subscriptions?area={GEO_POLYGON_STRING}")
            assert (
                resp.status_code == 200
            ), f"Expecting code 200, found {resp.status_code}"

            data = resp.json()
            found_deleted_sub = [
                sub["id"] for sub in data["subscriptions"] if sub["id"] in all_sub_1
            ]

            assert (
                found_deleted_sub == []
            ), f"Found deleted Subscriptions: {found_deleted_sub}"

    def testStep9(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Expired ISA automatically removed, ISA modifications
        accessible from all non-primary DSSs"""

        # sleep X seconds for ISA_1 to expire
        time.sleep(SHORT_WAIT_SEC)

        time_end = datetime.datetime.utcnow() + datetime.timedelta(
            seconds=SHORT_WAIT_SEC
        )
        for index, dss in enumerate([primary_dss] + all_other_dss):
            sub_2_uuid = str(uuid.uuid4())
            self.context[f"sub_2_{index}_uuid"] = TestContext("SUB", sub_2_uuid)
            resp = dss_map[dss].put(
                f"/subscriptions/{sub_2_uuid}",
                json={
                    "extents": {
                        "spatial_volume": {
                            "footprint": {"vertices": VERTICES},
                            "altitude_lo": 20,
                            "altitude_hi": 400,
                        },
                        "time_end": time_end.strftime(DATE_FORMAT),
                    },
                    "callbacks": {
                        "identification_service_area_url": "https://example.com/uss/identification_service_area"
                    },
                },
            )
            assert resp.status_code == 200, f"Failed to Insert Subscription to {dss}"
            data = resp.json()
            isa_ids = [isa["id"] for isa in data["service_areas"]]
            assert (
                self.context["isa_1_uuid"].uuid not in isa_ids
            ), f"{dss} returned expired ISA_1 when creating Subscription"

            # save SUB_2 Version String
            self.context[sub_2_uuid] = TestContext(
                "VERSION",
                data["subscription"]["version"],
            )


    def testStep10(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """ISA creation triggers subscription notification requests"""

        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
        self.context["isa_2_uuid"] = TestContext("ISA", str(uuid.uuid4()))
        resp = dss_map[primary_dss].put(
            f"/identification_service_areas/{self.context['isa_2_uuid'].uuid}",
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {"vertices": VERTICES},
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_end": time_end.strftime(DATE_FORMAT),
                },
                "flights_url": "https://example.com/uss/flights",
            },
        )
        assert (
            resp.status_code == 200
        ), f"Failed to insert ISA to {primary_dss}. Error: {resp.json()['error']}"

        all_sub_2 = set()
        for name, entity in self.context.items():
            if name.startswith("sub_2_"):
                all_sub_2.add(entity.uuid)

        returned_subs = self._extract_sub_ids_from_isa_put_response(resp.json())

        missing_subs = all_sub_2 - returned_subs
        assert (
            missing_subs == set()
        ), f"{primary_dss} returned too few Subscriptions, missing: {missing_subs}"

        # save ISA_2 Version String
        self.context[self.context["isa_2_uuid"].uuid] = TestContext(
            "VERSION",
            resp.json()["service_area"]["version"],
        )

    def testStep11(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """ISA deletion triggers subscription notification requests"""
        isa_2_uuid = self.context["isa_2_uuid"].uuid
        version = self.context[isa_2_uuid].uuid
        resp = dss_map[primary_dss].delete(
            f"/identification_service_areas/{isa_2_uuid}/{version}"
        )
        assert (
            resp.status_code == 200
        ), f"Failed to delete ISA to {primary_dss}. Error: {resp.json()['error']}"

        all_sub_2 = set()
        for name, entity in self.context.items():
            if name.startswith("sub_2_"):
                all_sub_2.add(entity.uuid)

        returned_subs = self._extract_sub_ids_from_isa_put_response(resp.json())

        missing_subs = all_sub_2 - returned_subs
        assert (
            missing_subs == set()
        ), f"{primary_dss} returned too few Subscriptions, missing: {missing_subs}"

    def testStep12(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """Expired Subscriptions don’t trigger subscription notification requests"""
        time.sleep(SHORT_WAIT_SEC)
        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
        self.context["isa_3_uuid"] = TestContext("ISA", str(uuid.uuid4()))
        resp = dss_map[primary_dss].put(
            f"/identification_service_areas/{self.context['isa_3_uuid'].uuid}",
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {"vertices": VERTICES},
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_end": time_end.strftime(DATE_FORMAT),
                },
                "flights_url": "https://example.com/uss/flights",
            },
        )
        assert (
            resp.status_code == 200
        ), f"Failed to insert ISA to {primary_dss}. Error: {resp.json()['error']}"

        all_sub_2 = [sub for sub in self.context if sub.startswith("sub_2_")]
        returned_subs = self._extract_sub_ids_from_isa_put_response(resp.json())
        found_expired_sub = [sub for sub in returned_subs if sub in all_sub_2]

        assert (
            found_expired_sub == []
        ), f"Found expired Subscriptions: {found_expired_sub}"

        # save ISA_3 Version String
        self.context[self.context["isa_3_uuid"].uuid] = TestContext(
            "VERSION",
            resp.json()["service_area"]["version"],
        )

    def testStep13(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Expired Subscription removed from geographic index on primary DSS"""
        all_dss = [primary_dss] + all_other_dss
        all_sub_2 = set()
        for index in range(len(all_dss)):
            all_sub_2.add(self.context[f"sub_2_{index}_uuid"].uuid)
        for index, dss in enumerate(all_dss):
            resp = dss_map[dss].get(f"/subscriptions?area={GEO_POLYGON_STRING}")
            assert resp.status_code == 200, f"{dss} failed to get SUB_2 by area"

            returned_subs = set([x["id"] for x in resp.json()["subscriptions"]])
            found_expired_sub = [sub for sub in returned_subs if sub in all_sub_2]

            assert (
                found_expired_sub == []
            ), f"Found expired Subscriptions: {found_expired_sub}"

    def testStep14(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Expired Subscription removed from ID index on primary DSS"""
        all_dss = [primary_dss] + all_other_dss
        for index, dss in enumerate(all_dss):
            sub_2_uuid = self.context[f"sub_2_{index}_uuid"].uuid
            resp = dss_map[dss].get(f"/subscriptions/{sub_2_uuid}")
            assert (
                resp.status_code == 404
            ), f"Expecting code 404, found {resp.status_code}"

    def testStep15(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """ISA deletion does not trigger subscription
        notification requests for expired Subscriptions"""
        isa_3_uuid = self.context["isa_3_uuid"].uuid
        version = self.context[isa_3_uuid].uuid
        resp = dss_map[primary_dss].delete(
            f"/identification_service_areas/{isa_3_uuid}/{version}"
        )
        assert resp.status_code == 200, f"Failed to delete ISA_3 from {primary_dss}"

        all_sub_2 = [sub for sub in self.context if sub.startswith("sub_2_")]
        returned_subs = self._extract_sub_ids_from_isa_put_response(resp.json())
        found_expired_sub = [sub for sub in returned_subs if sub in all_sub_2]

        assert (
            found_expired_sub == []
        ), f"Found expired Subscriptions: {found_expired_sub}"

    def testStep16(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Deleted ISA removed from all DSSs"""
        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
        for index, dss in enumerate([primary_dss] + all_other_dss):
            sub_3_uuid = str(uuid.uuid4())
            self.context[f"sub_3_{index}_uuid"] = TestContext("SUB", sub_3_uuid)
            resp = dss_map[dss].put(
                f"/subscriptions/{sub_3_uuid}",
                json={
                    "extents": {
                        "spatial_volume": {
                            "footprint": {"vertices": VERTICES},
                            "altitude_lo": 20,
                            "altitude_hi": 400,
                        },
                        "time_end": time_end.strftime(DATE_FORMAT),
                    },
                    "callbacks": {
                        "identification_service_area_url": "https://example.com/uss/identification_service_area"
                    },
                },
            )
            assert resp.status_code == 200, f"Failed to Insert Subscription to {dss}"
            data = resp.json()
            isa_ids = [isa["id"] for isa in data["service_areas"]]
            assert (
                self.context["isa_3_uuid"].uuid not in isa_ids
            ), f"{dss} returned deleted ISA_3 when creating Subscription"

            # save SUB_3 Version String
            self.context[sub_3_uuid] = TestContext(
                "VERSION",
                data["subscription"]["version"],
            )

    def testStep17(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """Clean up SUBS_3"""
        all_sub_3 = set()
        for name, entity in self.context.items():
            if name.startswith("sub_3_"):
                all_sub_3.add(entity.uuid)
        for sub_3_uuid in all_sub_3:
            version = self.context[sub_3_uuid].uuid
            resp = dss_map[primary_dss].delete(f"/subscriptions/{sub_3_uuid}/{version}")
            assert resp.status_code == 200, "Failed to delete SUB_3 from Primary DSS"
