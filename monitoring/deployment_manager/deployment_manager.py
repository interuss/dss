#!env/bin/python3

import argparse
import importlib
import json
import os
import pkgutil
import sys

from implicitdict import ImplicitDict
from monitoring.deployment_manager import actions
from monitoring.deployment_manager.systems.configuration import DeploymentSpec
from monitoring.deployment_manager.infrastructure import make_context
from monitoring.deployment_manager import infrastructure


def _import_submodules(module):
    for loader, module_name, is_pkg in pkgutil.walk_packages(module.__path__, module.__name__ + '.'):
        importlib.import_module(module_name)


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description='Manage an InterUSS deployment')

    actions = parser.add_subparsers(title='available deployment actions', dest='action', metavar='ACTION')
    for name, func in infrastructure.actions.items():
        actions.add_parser(name, help=func.__doc__)

    parser.add_argument('deployment_spec', metavar='SPEC', type=str, help='specification for deployment')

    return parser.parse_args()


def main() -> int:
    # Import all submodules from the `actions` module so we can find all actions
    _import_submodules(actions)

    # Parse arguments
    args = _parse_args()

    # Retrieve action function
    action_method = infrastructure.actions.get(args.action, None)
    if action_method is None:
        raise ValueError('Could not find definition for action `{}`'.format(args.action))

    # Parse deployment spec
    with open(args.deployment_spec, 'r') as f:
        spec = ImplicitDict.parse(json.load(f), DeploymentSpec)
    original_spec = json.dumps(spec)
    context = make_context(spec)

    # Execute action
    context.log.msg('Executing action', action=args.action, spec_file=args.deployment_spec)
    action_method(context)

    # Check if the deployment spec was updated
    new_spec = json.dumps(context.spec)
    if new_spec != original_spec:
        context.log.msg('Deployment spec updated; writing changes to {}'.format(args.deployment_spec))
        with open(args.deployment_spec, 'w') as f:
            json.dump(context.spec, f, indent=2)

    return os.EX_OK


if __name__ == '__main__':
    sys.exit(main())
