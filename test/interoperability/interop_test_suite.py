import clients
import itertools
import inspect
import datetime
import uuid
import collections
import time
import logging

from typing import Dict, Any, List

logging.basicConfig(level=logging.INFO)
LOG = logging.getLogger(__name__)

DATE_FORMAT = "%Y-%m-%dT%H:%M:%SZ"

VERTICES = [
    {"lng": 130.6205, "lat": -23.6558},
    {"lng": 130.6301, "lat": -23.6898},
    {"lng": 130.6700, "lat": -23.6709},
    {"lng": 130.6466, "lat": -23.6407},
]

GEO_POLYGON_STRING = ",".join("{},{}".format(x["lat"], x["lng"]) for x in VERTICES)

SHORT_WAIT_SEC = 5

TestContext = collections.namedtuple("TestContext", ["type", "key"])

class InterOpTestSuite:
    def __init__(self, dss_clients: Dict[str, clients.DSSClient]):
        self.dss_clients = dss_clients

    def startTest(self):
        round = 1
        for dss_permutation in itertools.permutations(self.dss_clients):
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
                        f"\tFailure Message: {e}"
                    )
                    LOG.error(msg)
                    LOG.debug(f"Cleaning up round {round}")
                    ts.cleanUp(self.dss_clients, primary_dss)
                    return (primary_dss, msg)

            LOG.debug(f"Cleaning up round {round}")
            ts.cleanUp(self.dss_clients, primary_dss)
            round += 1

    def _getTests(self) -> Dict[str, object]:
        methods = inspect.getmembers(TestSteps, predicate=inspect.isfunction)
        result: Dict[str, object] = {}
        ordered_result = collections.OrderedDict()
        for name, obj in methods:
            if name.startswith("testStep"):
                result[name] = obj
        # sort by name
        for i in range(1, len(result) + 1):
            name = f"testStep{i}"
            ordered_result[name] = result[name]

        return ordered_result


class TestSteps:
    def __init__(self):
        self.context: Dict[Any, TestContext] = {}

    def cleanUp(self, dss_map, primary_dss):
        dss = dss_map[primary_dss]
        for context_type, uuid in self.context.values():
            if context_type == "ISA":
                dss.delete(f"/identification_service_areas/{uuid}/")
            elif context_type == "SUB":
                dss.delete(f"/subscriptions/{uuid}/")
            else:
                LOG.warning(f"Unknown Type: {context_type}")

    def testStep1(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """Create ISA in Primary DSS with 10 min TTL."""

        time_start = datetime.datetime.utcnow()
        time_end = time_start + datetime.timedelta(minutes=10)
        self.context["isa_1_uuid"] = TestContext("ISA", str(uuid.uuid4()))
        resp = dss_map[primary_dss].put(
            f"/identification_service_areas/{self.context['isa_1_uuid'].key}",
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {"vertices": VERTICES},
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_start": time_start.strftime(DATE_FORMAT),
                    "time_end": time_end.strftime(DATE_FORMAT),
                },
                "flights_url": "https://example.com/dss",
            },
        )
        assert (
            resp.status_code == 200
        ), f"Failed to insert ISA to {primary_dss}. Error: {resp.json()['error']}"

    def testStep2(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Can create Subscription in all DSSs, ISA accessible from all
        non-primary DSSs."""
        time_start = datetime.datetime.utcnow()
        time_end = time_start + datetime.timedelta(minutes=10)
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
                        "time_start": time_start.strftime(DATE_FORMAT),
                        "time_end": time_end.strftime(DATE_FORMAT),
                    },
                    "callbacks": {
                        "identification_service_area_url": "https://example.com/foo"
                    },
                },
            )
            assert resp.status_code == 200, f"Failed to Insert Subscription to {dss}"
            data = resp.json()
            isa_ids = [isa["id"] for isa in data["service_areas"]]
            assert (
                self.context["isa_1_uuid"].key in isa_ids
            ), f"{dss} did not return ISA from testStep1 when creating Subscription"

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
                f"/subscriptions/{self.context['sub_1_0_uuid'].key}"
            )
            assert resp.status_code == 200, f"{dss} failed to get SUB_1"

            data = resp.json()
            assert (
                self.context["sub_1_0_uuid"].key == data["subscription"]["id"]
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
            all_sub_1.add(self.context[f"sub_1_{index}_uuid"].key)
        for index, dss in enumerate(all_dss):
            resp = dss_map[dss].get(f"/subscriptions?area={GEO_POLYGON_STRING}")
            assert resp.status_code == 200, f"{dss} failed to get SUB_1 by area"

            returned_subs = set([x["id"] for x in resp.json()["subscriptions"]])

            missing_subs = all_sub_1 - returned_subs
            assert (
                missing_subs == set()
            ), f"{dss} returned too few Subscriptions, missing: {missing_subs}"

            extra_subs = returned_subs - all_sub_1
            assert (
                extra_subs == set()
            ), f"{dss} returned too many Subscriptsions, extra: {extra_subs}"

    def testStep5(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """Can modify ISA in primary DSS, ISA modification triggers 
        subscription notification requests"""
        isa_1_uuid = self.context["isa_1_uuid"].key
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
                    "time_start": data["time_start"],
                    "time_end": time_end.strftime(DATE_FORMAT),
                },
                "flights_url": "https://example.com/dss",
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
        for name, context in self.context.items():
            if name.startswith("sub_1_"):
                resp = dss.delete(f"/subscriptions/{context.key}/")
                assert (
                    resp.status_code == 200
                ), f"Failed to delete {context.key} from Primary DSS: {primary_dss}"

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
            sub_uuid = self.context[f"sub_1_{index}_uuid"].key
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
        all_dss = [primary_dss] + all_other_dss
        for dss in all_dss:
            resp = dss_map[dss].get(f"/subscriptions?area={GEO_POLYGON_STRING}")
            assert (
                resp.status_code == 200
            ), f"Expecting code 200, found {resp.status_code}"

            data = resp.json()
            assert (
                data["subscriptions"] == []
            ), f"Expecting empty Subscription list, found {data['subscriptions']}"

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

        time_start = datetime.datetime.utcnow()
        time_end = time_start + datetime.timedelta(seconds=SHORT_WAIT_SEC)
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
                        "time_start": time_start.strftime(DATE_FORMAT),
                        "time_end": time_end.strftime(DATE_FORMAT),
                    },
                    "callbacks": {
                        "identification_service_area_url": "https://example.com/foo"
                    },
                },
            )
            assert resp.status_code == 200, f"Failed to Insert Subscription to {dss}"
            data = resp.json()
            isa_ids = [isa["id"] for isa in data["service_areas"]]
            assert (
                self.context["isa_1_uuid"].key not in isa_ids
            ), f"{dss} return ISA_1 when creating Subscription"

    def testStep10(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """ISA creation triggers subscription notification requests"""

        time_start = datetime.datetime.utcnow()
        time_end = time_start + datetime.timedelta(minutes=10)
        self.context["isa_2_uuid"] = TestContext("ISA", str(uuid.uuid4()))
        resp = dss_map[primary_dss].put(
            f"/identification_service_areas/{self.context['isa_2_uuid'].key}",
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {"vertices": VERTICES},
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_start": time_start.strftime(DATE_FORMAT),
                    "time_end": time_end.strftime(DATE_FORMAT),
                },
                "flights_url": "https://example.com/dss",
            },
        )
        assert (
            resp.status_code == 200
        ), f"Failed to insert ISA to {primary_dss}. Error: {resp.json()['error']}"

        all_sub_2 = set()
        for name, context in self.context.items():
            if name.startswith("sub_2_"):
                all_sub_2.add(context.key)

        returned_subs = set()
        for subscriber in resp.json()["subscribers"]:
            for subscription in subscriber:
                for sub in subscriber["subscriptions"]:
                    returned_subs.add(sub["subscription_id"])

        missing_subs = all_sub_2 - returned_subs
        assert (
            missing_subs == set()
        ), f"{dss} returned too few Subscriptions, missing: {missing_subs}"

        extra_subs = returned_subs - all_sub_2
        assert (
            extra_subs == set()
        ), f"{dss} returned too many Subscriptsions, extra: {extra_subs}"

    def testStep11(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """ISA deletion triggers subscription notification requests"""
        resp = dss_map[primary_dss].delete(
            f"/identification_service_areas/{self.context['isa_2_uuid'].key}/"
        )
        assert (
            resp.status_code == 200
        ), f"Failed to delete ISA to {primary_dss}. Error: {resp.json()['error']}"

        all_sub_2 = set()
        for name, context in self.context.items():
            if name.startswith("sub_2_"):
                all_sub_2.add(context.key)

        returned_subs = set()
        for subscriber in resp.json()["subscribers"]:
            for subscription in subscriber:
                for sub in subscriber["subscriptions"]:
                    returned_subs.add(sub["subscription_id"])

        missing_subs = all_sub_2 - returned_subs
        assert (
            missing_subs == set()
        ), f"{dss} returned too few Subscriptions, missing: {missing_subs}"

        extra_subs = returned_subs - all_sub_2
        assert (
            extra_subs == set()
        ), f"{dss} returned too many Subscriptsions, extra: {extra_subs}"

    def testStep12(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """Expired Subscriptions don’t trigger subscription notification requests"""
        time.sleep(SHORT_WAIT_SEC)
        time_start = datetime.datetime.utcnow()
        time_end = time_start + datetime.timedelta(minutes=10)
        self.context["isa_3_uuid"] = TestContext("ISA", str(uuid.uuid4()))
        resp = dss_map[primary_dss].put(
            f"/identification_service_areas/{self.context['isa_3_uuid'].key}",
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {"vertices": VERTICES},
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_start": time_start.strftime(DATE_FORMAT),
                    "time_end": time_end.strftime(DATE_FORMAT),
                },
                "flights_url": "https://example.com/dss",
            },
        )
        assert (
            resp.status_code == 200
        ), f"Failed to insert ISA to {primary_dss}. Error: {resp.json()['error']}"
        subscribers = resp.json()["subscribers"]

        assert (
            subscribers == []
        ), f"Expecting empty Subscribers list, found {len(subscribers)} subscribers"

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
            all_sub_2.add(self.context[f"sub_2_{index}_uuid"].key)
        for index, dss in enumerate(all_dss):
            resp = dss_map[dss].get(f"/subscriptions?area={GEO_POLYGON_STRING}")
            assert resp.status_code == 200, f"{dss} failed to get SUB_2 by area"

            subs = resp.json()["subscriptions"]
            assert subs == [], (
                f"Expecting empty Subscriptions list, "
                f"found {len(subs)} Subscriptions from {dss}"
            )

    def testStep14(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Expired Subscription removed from ID index on primary DSS"""
        all_dss = [primary_dss] + all_other_dss
        for index, dss in enumerate(all_dss):
            sub_2_uuid = self.context[f"sub_2_{index}_uuid"].key
            resp = dss_map[dss].get(f"/subscriptions/{sub_2_uuid}")
            assert (
                resp.status_code == 404
            ), f"Expecting code 404, found {resp.status_code}"

    def testStep15(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """ISA deletion triggers does not trigger subscription
        notification requests for expired Subscriptions"""
        isa_3_uuid = self.context["isa_3_uuid"].key
        resp = dss_map[primary_dss].delete(
            f"/identification_service_areas/{isa_3_uuid}/"
        )
        assert resp.status_code == 200, f"Failed to delete ISA_3 from {primary_dss}"

        subs = resp.json()["subscribers"]
        assert subs == [], (
            f"Expecting empty Subscribers list, " f"found {len(subs)} Subscribers"
        )

    def testStep16(
        self,
        dss_map: Dict[str, clients.DSSClient],
        primary_dss: str,
        all_other_dss: List[str],
    ) -> None:
        """Deleted ISA removed from all DSSs"""
        time_start = datetime.datetime.utcnow()
        time_end = time_start + datetime.timedelta(minutes=10)
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
                        "time_start": time_start.strftime(DATE_FORMAT),
                        "time_end": time_end.strftime(DATE_FORMAT),
                    },
                    "callbacks": {
                        "identification_service_area_url": "https://example.com/foo"
                    },
                },
            )
            assert resp.status_code == 200, f"Failed to Insert Subscription to {dss}"
            data = resp.json()
            isa_ids = [isa["id"] for isa in data["service_areas"]]
            assert (
                self.context["isa_3_uuid"].key not in isa_ids
            ), f"{dss} returned ISA_3 when creating Subscription"

    def testStep17(
        self, dss_map: Dict[str, clients.DSSClient], primary_dss: str, **kwargs
    ) -> None:
        """Clean up SUBS_3"""
        all_sub_3 = set()
        for name, context in self.context.items():
            if name.startswith("sub_3_"):
                all_sub_3.add(context.key)
        for sub_3_uuid in all_sub_3:
            resp = dss_map[primary_dss].delete(f"/subscriptions/{sub_3_uuid}/")
            assert resp.status_code == 200, "Failed to delete SUB_3 from Primary DSS"
