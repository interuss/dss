#!env/bin/python3

import datetime
import random
import threading
import time
import uuid
from locust import User, task, between
from monitoring.monitorlib import auth, infrastructure
from monitoring.prober.rid import common


class DSSClient(infrastructure.DSSTestSession):
    _locust_environment = None

    def request(self, method, url, **kwargs):
        if (method == "PUT" and len(url.split("/")) > 3) or method == "PATCH":
            real_method = "UPDATE"
        else:
            real_method = method
        name = url.split("/")[1]
        start_time = time.time()
        result = None
        try:
            result = super(DSSClient, self).request(method, url, **kwargs)
        except Exception as e:
            self.log_exception(real_method, name, start_time, e)
        else:
            if not result or result.status_code != 200:
                if not result:
                    msg = "Got None for Response"
                else:
                    msg = result.text
                self.log_exception(real_method, name, start_time, Exception(msg))
            else:
                total_time = int((time.time() - start_time) * 1000)
                self._locust_environment.events.request_success.fire(
                    request_type=real_method,
                    name=name,
                    response_time=total_time,
                    response_length=0,
                )
        return result

    def log_exception(self, real_method, name, start_time, e):
        total_time = int((time.time() - start_time) * 1000)
        self._locust_environment.events.request_failure.fire(
            request_type=real_method,
            name=name,
            response_time=total_time,
            exception=e,
            response_length=0,
        )


class USS(User):
    abstract = True
    isa_dict = {}
    sub_dict = {}

    def __init__(self, *args, **kwargs):
        super(USS, self).__init__(*args, **kwargs)
        oauth_adapter = auth.DummyOAuth("http://localhost:8085/token", "fake_uss")
        self.client = DSSClient(self.host, oauth_adapter)
        self.client._locust_environment = self.environment
        self.client.default_scopes = [common.SCOPE_WRITE, common.SCOPE_READ]


class ISA(USS):
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
                    "time_start": time_start.strftime(common.DATE_FORMAT),
                    "time_end": time_end.strftime(common.DATE_FORMAT),
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
                    "time_start": time_start.strftime(common.DATE_FORMAT),
                    "time_end": time_end.strftime(common.DATE_FORMAT),
                },
                "flights_url": "https://example.com/dss",
            },
        )
        if resp.status_code == 200:
            self.isa_dict[target_isa] = resp.json()["service_area"]["version"]

    @task(100)
    def get_isa(self):
        target_isa = random.choice(list(self.isa_dict.keys()))
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
        target_isa = random.choice(list(self.isa_dict.keys()))
        target_version = self.isa_dict.pop(target_isa, None)
        self.lock.release()
        return target_isa, target_version

    def on_start(self):
        # insert atleast 1 ISA for update to not fail
        self.create_isa()

    def on_stop(self):
        # clean up after itself
        for _ in self.isa_dict:
            self.delete_isa()


class SUB(USS):
    wait_time = between(0.01, 1)
    lock = threading.Lock()

    @task(100)
    def create_sub(self):
        time_start = datetime.datetime.utcnow()
        time_end = time_start + datetime.timedelta(minutes=60)
        sub_uuid = str(uuid.uuid4())

        resp = self.client.put(
            "/subscriptions/{}".format(sub_uuid),
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {
                            "vertices": common.VERTICES,
                        },
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_start": time_start.strftime(common.DATE_FORMAT),
                    "time_end": time_end.strftime(common.DATE_FORMAT),
                },
                "callbacks": {
                    "identification_service_area_url": "https://example.com/foo"
                },
            },
        )
        if resp.status_code == 200:
            self.sub_dict[sub_uuid] = resp.json()["service_area"]["version"]

    @task(20)
    def get_sub(self):
        target_sub = random.choice(list(self.sub_dict.keys()))
        if not target_sub:
            print("Nothing to pick from sub_dict for GET")
            return
        self.client.get("/subscriptions/{}".format(target_sub))

    @task(50)
    def update_sub(self):
        target_sub, target_version = self.checkout_sub()
        if not target_sub:
            print("Nothing to pick from sub_dict for UPDATE")
            return

        time_start = datetime.datetime.utcnow()
        time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=2)
        resp = self.client.put(
            "/subscriptions/{}/{}".format(target_sub, target_version),
            json={
                "extents": {
                    "spatial_volume": {
                        "footprint": {
                            "vertices": common.VERTICES,
                        },
                        "altitude_lo": 20,
                        "altitude_hi": 400,
                    },
                    "time_start": time_start.strftime(common.DATE_FORMAT),
                    "time_end": time_end.strftime(common.DATE_FORMAT),
                },
                "callbacks": {
                    "identification_service_area_url": "https://example.com/foo"
                },
            },
        )
        if resp.status_code == 200:
            self.sub_dict[target_sub] = resp.json()["service_area"]["version"]

    @task(5)
    def delete_sub(self):
        target_sub, target_version = self.checkout_sub()
        if not target_sub:
            print("Nothing to pick from sub_dict for DELETE")
            return
        self.client.delete("/subscriptions/{}/{}".format(target_sub, target_version))

    def checkout_sub(self):
        self.lock.acquire()
        target_sub = random.choice(list(self.sub_dict.keys()))
        target_version = self.sub_dict.pop(target_sub, None)
        self.lock.release()
        return target_sub, target_version

    def on_start(self):
        # Insert atleast 1 Sub for update to not fail
        self.create_sub()

    def on_stop(self):
        # clean up after itself
        for _ in self.sub_dict:
            self.delete_sub()
