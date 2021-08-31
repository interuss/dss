import functools

import pytest


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
