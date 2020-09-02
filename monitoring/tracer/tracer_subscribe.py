#!env/bin/python3

import argparse
import datetime
import os
import sys
from typing import Dict

import s2sphere

from monitoring.monitorlib import auth, ids, infrastructure, rid
from . import geo, tracerlog


RID_SUBSCRIPTION_ID_CODE = 'tracer RID Subscription'
SCD_SUBSCRIPTION_ID_CODE = 'tracer SCD Subscription'


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Test Interoperability of DSSs")

    # Required arguments
    parser.add_argument('--auth', help='Auth spec for obtaining authorization to DSS and USSs; see README.md')
    parser.add_argument('--dss', help='Base URL of DSS instance to query')
    parser.add_argument('--area', help='`lat,lng,lat,lng` for box containing the area to trace interactions for')
    parser.add_argument('--output-folder', help='Path of folder in which to write logs')

    # Feature arguments
    parser.add_argument('--rid-subscription-callback', default=None, help='Base URL of implementation of USS-USS RID callback for ISA-sensitive Subscriptions')
    parser.add_argument('--scd-subscription-callback', default=None, help='Base URL of implementation of USS-USS SCD callback for Operation- and Constraint-sensitive Subscriptions')

    return parser.parse_args()


def main() -> int:
    args = parseArgs()

    # Required resources
    adapter: auth.AuthAdapter = auth.make_auth_adapter(args.auth)
    dss_client = infrastructure.DSSTestSession(args.dss, adapter)
    area: s2sphere.LatLngRect = geo.make_latlng_rect(args.area)
    logger: tracerlog.Logger = tracerlog.Logger(args.output_folder)
    uss_clients: Dict[str, infrastructure.DSSTestSession] = {}

    # RID Subscription
    if args.rid_subscription_callback is not None:
      sub_id = ids.make_id(RID_SUBSCRIPTION_ID_CODE)
      time_start = datetime.datetime.utcnow()
      time_end = time_start + datetime.timedelta(hours=18)

      resp = dss_client.put(
        '/subscriptions/{}'.format(sub_id),
        json={
          'extents': {
            'spatial_volume': {
              'footprint': {
                'vertices': common.VERTICES,
              },
              'altitude_lo': 0,
              'altitude_hi': 3048,
            },
            'time_start': time_start.strftime(rid.DATE_FORMAT),
            'time_end': time_end.strftime(rid.DATE_FORMAT),
          },
          'callbacks': {
            'identification_service_area_url': args.rid_subscription_callback
          },
        })
      assert resp.status_code == 200

    # SCD Subscription
    if args.scd_subscription_callback is not None:
      raise NotImplementedError('tracer does not yet support SCD Subscription creation')




    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
