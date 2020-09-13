#!env/bin/python3

import argparse
import datetime
import os
import sys
import time
from typing import List

from monitoring.monitorlib import versioning
from monitoring.tracer import formatting, polling
from monitoring.tracer.resources import ResourceSet


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Poll for changes in DSS-tracked Entity status")

    ResourceSet.add_arguments(parser)

    # Feature arguments
    parser.add_argument('--rid-isa-poll-interval', type=float, default=0, help='Seconds beteween each poll of the DSS for ISAs, 0 to disable DSS polling for ISAs')
    parser.add_argument('--scd-operation-poll-interval', type=float, default=0, help='Seconds between each poll of the DSS for Operations, 0 to disable DSS polling for Operations')
    parser.add_argument('--scd-constraint-poll-interval', type=float, default=0, help='Seconds between each poll of the DSS for Constraints, 0 to disable DSS polling for Constraints')

    return parser.parse_args()


def print_no_newline(s):
  sys.stdout.write(s)
  sys.stdout.flush()


def main() -> int:
    args = parseArgs()

    # Required resources
    resources = ResourceSet.from_arguments(args)

    config = vars(args)
    config['code_version'] = versioning.get_code_version()
    resources.logger.log_new('poll_start', config)

    # Prepare pollers
    pollers: List[polling.Poller] = []

    if args.rid_isa_poll_interval > 0:
      pollers.append(polling.Poller(
        name='poll_isas',
        object_diff_text=formatting.isa_diff_text,
        interval=datetime.timedelta(seconds=args.rid_isa_poll_interval),
        poll=lambda: polling.poll_rid_isas(resources, resources.area)))

    if args.scd_operation_poll_interval > 0:
      pollers.append(polling.Poller(
        name='poll_ops',
        object_diff_text=formatting.entity_diff_text,
        interval=datetime.timedelta(seconds=args.scd_operation_poll_interval),
        poll=lambda: polling.poll_scd_operations(resources)))

    if args.scd_constraint_poll_interval > 0:
      pollers.append(polling.Poller(
        name='poll_constraints',
        object_diff_text=formatting.entity_diff_text,
        interval=datetime.timedelta(seconds=args.scd_constraint_poll_interval),
        poll=lambda: polling.poll_scd_constraints(resources)))

    if len(pollers) == 0:
      sys.stderr.write('Bad arguments: No data types had polling requests')
      return os.EX_USAGE

    # Execute the polling loop
    abort = False
    need_line_break = False
    while not abort:
      try:
        most_urgent_dt = datetime.timedelta(days=999999999)
        most_urgent_poller = None
        for poller in pollers:
          dt = poller.time_to_next_poll()
          if dt < most_urgent_dt:
            most_urgent_poller = poller
            most_urgent_dt = dt

        if most_urgent_dt.total_seconds() > 0:
          time.sleep(most_urgent_dt.total_seconds())

        result = most_urgent_poller.poll()

        if result.has_different_content_than(most_urgent_poller.last_result):
          resources.logger.log_new(most_urgent_poller.name, result.to_json())
          if need_line_break:
            print()
          print(most_urgent_poller.diff_text(result))
          need_line_break = False
          most_urgent_poller.last_result = result
        else:
          resources.logger.log_same(result.initiated_at, result.completed_at, most_urgent_poller.name)
          print_no_newline('.')
          need_line_break = True
      except KeyboardInterrupt:
        abort = True

    resources.logger.log_new('poll_stop', {
      'timestamp': datetime.datetime.utcnow().isoformat(),
    })

    return os.EX_OK

if __name__ == "__main__":
    sys.exit(main())
