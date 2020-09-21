#!env/bin/python3

import client
import datetime
import random
import threading
import time
import typing
import uuid
from monitoring.monitorlib import rid
from monitoring.prober.rid import common
from locust import task, between

class ISA(client.USS):
    wait_time = between(0.01, 1)
    lock = threading.Lock()

    @task(10)
    def create_isa(self):
        time_start = datetime.datetime.utcnow()
        time_end = time_start + datetime.timedelta(minutes=60)
        isa_uuid = str(uuid.uuid4())

        resp = self.client.put(
            "/identification_service_areas/{}".format(isa_uuid),
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {
                            "vertices": common.VERTICES,
                        },
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_start": time_start.strftime(rid.DATE_FORMAT),
                    "time_end": time_end.strftime(rid.DATE_FORMAT),
                },
                "flights_url": "https://example.com/dss",
            },
        )
        if resp.status_code == 200:
            self.isa_dict[isa_uuid] = resp.json()["service_area"]["version"]

    @task(5)
    def update_isa(self):
        target_isa, target_version = self.checkout_isa()
        if not target_isa:
            print("Nothing to pick from isa_dict for UPDATE")
            return

        time_start = datetime.datetime.utcnow()
        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=2)
        resp = self.client.put(
            "/identification_service_areas/{}/{}".format(target_isa, target_version),
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {
                            "vertices": common.VERTICES,
                        },
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_start": time_start.strftime(rid.DATE_FORMAT),
                    "time_end": time_end.strftime(rid.DATE_FORMAT),
                },
                "flights_url": "https://example.com/dss",
            },
        )
        if resp.status_code == 200:
            self.isa_dict[target_isa] = resp.json()["service_area"]["version"]

    @task(100)
    def get_isa(self):
        target_isa = random.choice(list(self.isa_dict.keys())) if self.isa_dict else None
        if not target_isa:
            print("Nothing to pick from isa_dict for GET")
            return
        self.client.get("/identification_service_areas/{}".format(target_isa))

    @task(1)
    def delete_isa(self):
        target_isa, target_version = self.checkout_isa()
        if not target_isa:
            print("Nothing to pick from isa_dict for DELETE")
            return
        self.client.delete(
            "/identification_service_areas/{}/{}".format(target_isa, target_version)
        )

    def checkout_isa(self):
        self.lock.acquire()
        target_isa = random.choice(list(self.isa_dict.keys())) if self.isa_dict else None
        target_version = self.isa_dict.pop(target_isa, None)
        self.lock.release()
        return target_isa, target_version

    def on_start(self):
        # insert atleast 1 ISA for update to not fail
        self.create_isa()

    def on_stop(self):
        self.isa_dict = {}

