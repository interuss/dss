import requests
import time
from enum import Enum
from google.auth.transport import requests as google_requests
from google.oauth2 import service_account
from typing import Optional, Dict, List, Any
import urllib


class AuthType(Enum):
    NONE = 0
    SVC_ACC = 1
    PASSWORD = 2


SCOPES = [
    "dss.write.identification_service_areas",
    "dss.read.identification_service_areas",
]


class OAuthClient:
    def __init__(
        self,
        endpoint: str,
        auth_type: AuthType,
        service_account_json: Optional[str] = "",
        username: Optional[str] = "",
        password: Optional[str] = "",
        client_id: Optional[str] = "",
    ):
        self._endpoint = endpoint
        self._token_cache = {}
        self._req_params = {}
        self.req = requests.Session()

        if auth_type is AuthType.SVC_ACC:
            credentials = service_account.Credentials.from_service_account_file(
                service_account_json
            ).with_scopes(["email"])
            self.req = google_requests.AuthorizedSession(credentials)
            self._req_params = {"grant_type": "client_credentials"}
        elif auth_type is AuthType.PASSWORD:
            self._req_params = {
                "grant_type": "password",
                "username": username,
                "password": password,
                "client_id": client_id,
            }
        self.parameterized_url = False

    def _isExpired(self, token: Dict[str, Any]) -> bool:
        expiration = token.get("expire_time")
        if not expiration:
            return False
        return expiration > time.time()

    def getToken(self, scopes_list: List[str], audience: str) -> str:
        scopes = " ".join(scopes_list)
        token = self._token_cache.get((scopes, audience))
        if token is None or self._isExpired(token):
            self._token_cache[(scopes, audience)] = self._issueToken(scopes, audience)
        return self._token_cache[(scopes, audience)].get("access_token", "")

    def _issueToken(self, scopes, audience) -> Dict[str, Any]:
        data = {"scope": scopes, "intended_audience": audience}
        data.update(self._req_params)
        if self.parameterized_url:
            param_str = "?" + "&".join([f"{key}={val}" for key, val in data.items()])
            response = self.req.post((self._endpoint + param_str))
        else:
            response = self.req.post(self._endpoint, data=data)
        response.raise_for_status()
        return response.json()


class DSSClient(requests.Session):
    def __init__(self, host: str, oauth_client: OAuthClient):
        super().__init__()
        self._host = host
        self._oauth_client = oauth_client
        self.scope: List[str] = []
        self.intended_audience: str = ""
        self.scope: List[str] = [
            "dss.write.identification_service_areas",
            "dss.read.identification_service_areas",
        ]
        self.intended_audience = urllib.parse.urlparse(host).hostname

    def prepare_request(self, request, **kwargs) -> requests.request:
        token = self._oauth_client.getToken(self.scope, self.intended_audience)
        if request.url.startswith("/"):
            request.url = self._host + request.url
        request.headers["Authorization"] = f"Bearer {token}"
        return super().prepare_request(request, **kwargs)
