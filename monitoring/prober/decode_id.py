import argparse
import importlib
import pkgutil
import sys
from typing import List

import monitoring
from monitoring.prober import infrastructure


# Import all tests
def import_submodules(package, recursive=True):
  """ Import all submodules of a module, recursively, including subpackages

  :param package: package (name or actual module)
  :type package: str | module
  :rtype: dict[str, types.ModuleType]
  """
  if isinstance(package, str):
    package = importlib.import_module(package)
  results = {}
  for loader, name, is_pkg in pkgutil.walk_packages(package.__path__):
    full_name = package.__name__ + '.' + name
    results[full_name] = importlib.import_module(full_name)
    if recursive and is_pkg:
      results.update(import_submodules(full_name))
  return results

import_submodules(monitoring.prober)


def parse_args(argv: List[str]):
  parser = argparse.ArgumentParser(description='Decode a test ID')
  parser.add_argument(action='store', dest='id', type=str, metavar='ID',
    help='The ID to decode')
  return parser.parse_args(argv)


if __name__ == '__main__':
  args = parse_args(sys.argv[1:])
  owner_name, id_code = infrastructure.IDFactory.decode(args.id)
  print('Owner: {}'.format(owner_name))
  if id_code in infrastructure.resource_type_code_descriptions:
    print('Resource type: {}'.format(infrastructure.resource_type_code_descriptions[id_code]))
  else:
    print('Resource type: Unregistered ({})'.format(id_code))
