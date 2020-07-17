import base64
import datetime
import re
from typing import List, Optional
import urllib.parse

import cryptography.exceptions
import cryptography.hazmat.backends
import cryptography.hazmat.primitives.hashes
import cryptography.hazmat.primitives.serialization
import cryptography.x509
import jwcrypto.common
import jwcrypto.jwk
import jwcrypto.jws
import requests
from google.auth.transport import requests as google_requests
from google.oauth2 import service_account

from .infrastructure import AuthAdapter


class DummyOAuth(AuthAdapter):
  """Auth adapter that gets JWTs that uses the Dummy OAuth Server"""

  def __init__(self, token_endpoint: str, sub: str):
    super().__init__()

    self._oauth_token_endpoint = token_endpoint
    self._sub = sub
    self._oauth_session = requests.Session()

  # Overrides method in AuthAdapter
  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    url = '{}?grant_type=client_credentials&scope={}&intended_audience={}&issuer=dummy&sub={}'.format(
      self._oauth_token_endpoint, urllib.parse.quote(' '.join(scopes)),
      urllib.parse.quote(intended_audience), self._sub)
    response = self._oauth_session.post(url).json()
    return response['access_token']


class ServiceAccount(AuthAdapter):
  """Auth adapter that gets JWTs using a service account."""

  def __init__(self, token_endpoint: str, service_account_json: str):
    super().__init__()

    credentials = service_account.Credentials.from_service_account_file(
      service_account_json).with_scopes(['email'])
    oauth_session = google_requests.AuthorizedSession(credentials)

    self._oauth_token_endpoint = token_endpoint
    self._oauth_session = oauth_session

  # Overrides method in AuthAdapter
  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    url = '{}?grant_type=client_credentials&scope={}&intended_audience={}'.format(
      self._oauth_token_endpoint, urllib.parse.quote(' '.join(scopes)),
      urllib.parse.quote(intended_audience))
    response = self._oauth_session.post(url).json()
    return response['access_token']


class UsernamePassword(AuthAdapter):
  """Auth adapter that gets JWTs using a username and password."""

  def __init__(self, token_endpoint: str, username: str, password: str, client_id: str):
    super().__init__()

    self._oauth_token_endpoint = token_endpoint
    self._username = username
    self._password = password
    self._client_id = client_id

  # Overrides method in AuthAdapter
  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    scopes.append('aud:{}'.format(intended_audience))
    response = requests.post(self._oauth_token_endpoint, data={
      'grant_type': "password",
      'username': self._username,
      'password': self._password,
      'client_id': self._client_id,
      'scope': ' '.join(scopes),
    }).json()
    return response['access_token']


class SignedRequest(AuthAdapter):
  """Auth adapter that gets JWTs by signing its outgoing requests."""

  def __init__(self, token_endpoint: str, client_id: str, key_path: str, cert_url: str, key_id: Optional[str] = None):
    """Create an AuthAdapter that retrieves tokens via message signing.

    Args:
      token_endpoint: URL of the authorization server's token endpoint.
      client_id: ID of client for which the token is being requested.
      key_path: Path to private key with which to sign the token request.
      cert_url: Publicly-accessible URL of certificate containing the public key
        corresponding to the private key in key_path and signed by an authority
        recognized by the authorization server.
      key_id: If specified, the specific ID to supply in the JWS header.  If not
        specified, defaults to the thumbprint of the certificate's public key.
    """
    super().__init__()

    self._token_endpoint = token_endpoint
    self._client_id = client_id
    with open(key_path, 'r') as f:
      self._private_key = f.read()
    self._cert_url = cert_url
    self._backend = cryptography.hazmat.backends.default_backend()

    # Retrieve certificate to validate match with private key
    response = requests.get(self._cert_url)
    assert response.status_code == 200
    if cert_url[-4:].lower() == '.der':
      cert = cryptography.x509.load_der_x509_certificate(response.content, self._backend)
    elif cert_url[-4:].lower() == '.crt':
      cert = cryptography.x509.load_pem_x509_certificate(response.content, self._backend)
    else:
      raise ValueError('cert_url must end with .der or .crt')
    cert_public_key = cert.public_key().public_bytes(
      cryptography.hazmat.primitives.serialization.Encoding.PEM,
      cryptography.hazmat.primitives.serialization.PublicFormat.SubjectPublicKeyInfo)

    # Generate public key directly from private key
    with open(key_path, 'r') as f:
      key_content = f.read().encode('utf-8')
    if key_path[-4:].lower() == '.key' or key_path[-4:].lower() == '.pem':
      private_key = cryptography.hazmat.primitives.serialization.load_pem_private_key(
        key_content, password=None, backend=self._backend)
      private_key_bytes = key_content
    else:
      raise ValueError('key_path must end with .key or .pem')
    public_key = private_key.public_key().public_bytes(
      cryptography.hazmat.primitives.serialization.Encoding.PEM,
      cryptography.hazmat.primitives.serialization.PublicFormat.SubjectPublicKeyInfo)

    if cert_public_key != public_key:
      raise ValueError('Public key in certificate does not match private key provided')

    self._private_jwk = jwcrypto.jwk.JWK.from_pem(private_key_bytes)
    self._public_jwk = jwcrypto.jwk.JWK.from_pem(public_key)

    # Assign key ID
    if key_id:
      self._kid = key_id
    else:
      self._kid = self._public_jwk.thumbprint()

  # Overrides method in AuthAdapter
  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    # Construct request body
    query = {
      'grant_type': 'client_credentials',
      'client_id': self._client_id,
      'scope': ' '.join(scopes),
      'current_timestamp': datetime.datetime.utcnow().isoformat() + 'Z',
    }
    payload = '&'.join([k + '=' + v for k, v in query.items()])

    # Construct JWS
    token_headers = {
      'typ': 'JOSE',
      'alg': 'RS256',
      'x5u': self._cert_url,
      'kid': self._kid,
    }
    jws = jwcrypto.jws.JWS(payload.encode('utf-8'))
    jws.add_signature(self._private_jwk, 'RS256', protected=jwcrypto.common.json_encode(token_headers))
    signed = jws.serialize(compact=True)

    # Check JWS
    jws_check = jwcrypto.jws.JWS()
    jws_check.deserialize(signed)
    try:
      jws_check.verify(self._public_jwk, 'RS256')
    except jwcrypto.jws.InvalidJWSSignature:
      raise ValueError('Could not construct a valid cryptographic signature for JWS')

    # Construct signature
    signature = re.sub(r'\.[^.]*\.', '..', signed)

    # Make token request
    request_headers = {
      'Content-Type': 'application/x-www-form-urlencoded',
      'x-utm-message-signature': signature,
    }
    response = requests.post(self._token_endpoint, data=payload, headers=request_headers)
    if response.status_code != 200:
      raise ValueError('Unable to retrieve access token:\n' + response.content.decode('utf-8'))
    return response.json()['access_token']


class ClientIdClientSecret(AuthAdapter):
  """Auth adapter that gets JWTs using a client ID and client secret."""

  def __init__(self, token_endpoint: str, client_id: str, client_secret: str):
    super().__init__()

    self._oauth_token_endpoint = token_endpoint
    self._client_id = client_id
    self._client_secret = client_secret

  # Overrides method in AuthAdapter
  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    response = requests.post(self._oauth_token_endpoint, json={
      'grant_type': 'client_credentials',
      'client_id': self._client_id,
      'client_secret': self._client_secret,
      'audience': intended_audience,
      'scope': ' '.join(scopes),
    })
    if response.status_code != 200:
      raise ValueError('Unable to retrieve access token:\n' + response.content.decode('utf-8'))
    return response.json()['access_token']


def make_auth_adapter(spec: str) -> AuthAdapter:
  """Make an AuthAdapter according to a string specification.

  Args:
    spec: Specification of adapter in the form
      ADAPTER_NAME([VALUE1[,PARAM2=VALUE2][,...]]) where ADAPTER_NAME is the
      name of a subclass of AuthAdapter and the contents of the parentheses are
      *args-style and **kwargs-style values for the parameters of ADAPTER_NAME's
      __init__, but the values (all strings) do not have any quote-like
      delimiters.

  Returns:
    An instance of the appropriate AuthAdapter subclass according to the
    provided spec.
  """
  m = re.match(r'^\s*([^\s(]+)\s*\(\s*([^)]*)\s*\)\s*$', spec)
  if m is None:
    raise ValueError('Auth adapter specification did not match the pattern `AdapterName(param, param, ...)`')

  adapter_name = m.group(1)
  adapter_classes = {cls.__name__: cls for cls in AuthAdapter.__subclasses__()}
  if adapter_name not in adapter_classes:
    raise ValueError('Auth adapter `%s` does not exist' % adapter_name)
  Adapter = adapter_classes[adapter_name]

  adapter_param_string = m.group(2)
  param_strings = [s.strip() for s in adapter_param_string.split(',')]
  args = []
  kwargs = {}
  for param_string in param_strings:
    if '=' in param_string:
      kv = param_string.split('=')
      if len(kv) != 2:
        raise ValueError('Auth adapter specification contained a parameter with more than one `=` character')
      kwargs[kv[0].strip()] = kv[1].strip()
    else:
      args.append(param_string)

  return Adapter(*args, **kwargs)
