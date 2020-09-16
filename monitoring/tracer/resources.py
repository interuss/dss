import argparse
import datetime

import s2sphere

from monitoring.monitorlib import auth, infrastructure, geo
from monitoring.tracer import tracerlog


class ResourceSet(object):
  """Set of resources necessary to obtain information from the UTM system."""

  def __init__(self,
               dss_client: infrastructure.DSSTestSession,
               area: s2sphere.LatLngRect,
               logger: tracerlog.Logger,
               start_time: datetime.datetime,
               end_time: datetime.datetime):
    self.dss_client = dss_client
    self.area = area
    self.logger = logger
    self.start_time = start_time
    self.end_time = end_time

    self.scd_cache = {}

  @classmethod
  def add_arguments(cls, parser: argparse.ArgumentParser) -> None:
    parser.add_argument('--auth', required=True, help='Auth spec for obtaining authorization to DSS and USSs; see README.md')
    parser.add_argument('--dss', required=True, help='Base URL of DSS instance to query')
    parser.add_argument('--area', required=True, help='`lat,lng,lat,lng` for box containing the area to trace interactions for')
    parser.add_argument('--start-time', default=datetime.datetime.utcnow().isoformat(), help='ISO8601 UTC datetime at which to start polling')
    parser.add_argument('--trace-hours', type=float, default=18, help='Number of hours to trace for')
    parser.add_argument('--output-folder', help='Path of folder in which to write logs')

  @classmethod
  def from_arguments(cls, args: argparse.Namespace):
    adapter: auth.AuthAdapter = auth.make_auth_adapter(args.auth)
    dss_client = infrastructure.DSSTestSession(args.dss, adapter)
    area: s2sphere.LatLngRect = geo.make_latlng_rect(args.area)
    start_time = datetime.datetime.fromisoformat(args.start_time)
    end_time = start_time + datetime.timedelta(hours=args.trace_hours)
    logger = tracerlog.Logger(args.output_folder) if args.output_folder else None
    return ResourceSet(dss_client, area, logger, start_time, end_time)
