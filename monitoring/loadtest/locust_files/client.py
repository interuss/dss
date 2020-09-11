#!env/bin/python3

import conf
import time
import typing
from locust import User 
from monitoring.monitorlib import auth, infrastructure
from monitoring.prober.rid import common

class DSSClient(infrastructure.DSSTestSession):
    _locust_environment = None

    def request(self, method: str, url: str, **kwargs):
        if (method == "PUT" and len(url.split("/")) > 3) or method == "PATCH":
            real_method = "UPDATE"
        else:
            real_method = method
        name = url.split("/")[1]
        start_time = time.time()
        result = None
        try:
            result = super().request(method, url, **kwargs)
        except Exception as e:
            self.log_exception(real_method, name, start_time, e)
        else:
            if result is None or result.status_code != 200:
                if result is None:
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

    def log_exception(self, real_method: str, name: str, start_time: float, e: Exception):
        total_time = int((time.time() - start_time) * 1000)
        self._locust_environment.events.request_failure.fire(
            request_type=real_method,
            name=name,
            response_time=total_time,
            exception=e,
            response_length=0,
        )

class USS(User):
    # Suggested by Locust 1.2.2 API Docs https://docs.locust.io/en/stable/api.html#locust.User.abstract
    abstract = True
    isa_dict: typing.Dict[str, str] = {}
    sub_dict: typing.Dict[str, str] = {}

    def __init__(self, *args, **kwargs):
        super(USS, self).__init__(*args, **kwargs)
        oauth_adapter = auth.DummyOAuth(conf.OAUTH_HOST, "fake_uss")
        self.client = DSSClient(self.host, oauth_adapter)
        self.client._locust_environment = self.environment
        # This is a load tester its acceptable to have all the scopes required to operate anything.
        # We are not testing if the scope is incorrect. We are testing if it can handle the load.
        self.client.default_scopes = [common.SCOPE_WRITE, common.SCOPE_READ]