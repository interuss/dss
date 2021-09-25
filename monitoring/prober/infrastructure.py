import functools
import inspect
import os
from typing import Dict, Tuple

import pytest

from monitoring.prober import utils


class VersionString(str):
  pass


def for_api_versions(*args):
  """Test decorator that checks if API version being tested applies to the test.

  A test function decorated with this decorator must include an argument that
  matches a fixture which takes on a VersionString value.  If that VersionString
  value matches any of the values specified in this decorator, the test will
  proceed normally.  If it does not match any of the values specified in this
  decorator, the test will be skipped.

  :param args: List of API versions for which the decorated test should be run.
  """
  def decorator_default_scope(func):
    acceptable_versions = args
    @functools.wraps(func)
    def wrapper_default_scope(*args, **kwargs):
      api_version = None
      for arg in args:
        if isinstance(arg, VersionString):
          api_version = arg
          break
      for key, value in kwargs.items():
        if isinstance(value, VersionString):
          api_version = value
          break

      if api_version is None:
        raise ValueError('A test with the @for_api_versions decorator must include, in its arguments, a fixture populated with a VersionString value (for instance: scd_api)')
      if api_version in acceptable_versions:
        return func(*args, **kwargs)
      else:
        pytest.skip('Not applicable for API version {}'.format(api_version))

    return wrapper_default_scope
  return decorator_default_scope


resource_type_code_descriptions: Dict[int, str] = {}


# Next resource_type_code: 2
def register_resource_type(code: int, description: str) -> int:
  test_filename = inspect.stack()[1].filename
  this_folder = os.path.dirname(os.path.abspath(__file__))
  test = test_filename[len(this_folder)+1:]
  full_description = '{}: {}'.format(test, description)
  if code in resource_type_code_descriptions:
    raise ValueError('Resource type code {} is already in use as "{}" so it cannot be used for "{}"'.format(code, resource_type_code_descriptions[code], full_description))
  resource_type_code_descriptions[code] = full_description
  return code


class IDFactory(object):
  """Creates UUIDv4-formatted IDs encoding the kind of ID and owner.

  Format: 0000XXXX-YYYY-4ZYY-YYYY-YYYYYYYY0000

  XXXX encodes the kind of ID according to id_codes.
  YYYYYYYYYYYYYYYYY encodes the owner/creator of the resource having the ID and
  consists of 24 characters encoded as 6-bit groups.
  ZZ is reserved and currently set to 0.
  """

  owner_id: str

  def __init__(self, test_owner: str):
    self.owner_id = utils.encode_owner(test_owner)

  def make_id(self, resource_type: int):
    """Make a test ID with the specified resource type code"""
    return '0000{x}-{y1}-40{y2}-{y3}-{y4}0000'.format(
      x=utils.encode_resource_type_code(resource_type),
      y1=self.owner_id[0:4],
      y2=self.owner_id[4:6],
      y3=self.owner_id[6:10],
      y4=self.owner_id[10:18])

  @classmethod
  def decode(cls, id: str) -> Tuple[str, int]:
    hex_digits = id.replace('-', '')
    if len(hex_digits) != 32:
      raise ValueError('ID {} has the wrong number of characters for a UUID'.format(id))
    if hex_digits[0:4] != '0000' or hex_digits[-4:] != '0000':
      raise ValueError('ID {} does not have the leading and trailing zeros indicating a test ID'.format(id))
    if hex_digits[12:14] != '40':
      raise ValueError('ID {} is not formatted like a v4 test ID'.format(id))
    x = hex_digits[4:8]
    y = hex_digits[8:12] + hex_digits[14:28]
    resource_type_code = int(x, 16)
    owner_name = utils.decode_owner(y)
    return owner_name, resource_type_code
