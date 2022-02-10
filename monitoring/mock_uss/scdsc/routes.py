from typing import List, Tuple

import flask

from monitoring.monitorlib import scd
from monitoring.monitorlib.clients import scd as scd_client
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.mock_uss import resources, webapp
from monitoring.mock_uss.scdsc.database import db


@webapp.route('/scdsc/status')
def scdsc_status():
    return 'Mock SCD strategic coordinator ok'


@webapp.route('/scdsc/clear_all_flights', methods=['POST'])
def clear_all_flights() -> Tuple[str, int]:
    """Clear all flights from the system that this mock USS may have been involved with.

    Expected request body is a JSON form of scd.DeleteAllFlightsRequest.
    """
    try:
        json = flask.request.json
        if json is None:
            raise ValueError('Request did not contain a JSON payload')
        req = ImplicitDict.parse(json, scd.DeleteAllFlightsRequest)
    except ValueError as e:
        msg = 'Unable to parse DeleteAllFlightsRequest JSON request: {}'.format(e)
        return msg, 400

    # Find operational intents in the DSS
    start_time = scd.start_of(req.extents)
    end_time = scd.end_of(req.extents)
    area = scd.rect_bounds_of(req.extents)
    alt_lo, alt_hi = scd.meter_altitude_bounds_of(req.extents)
    vol4 = scd.make_vol4(start_time, end_time, alt_lo, alt_hi, polygon=scd.make_polygon(latlngrect=area))
    try:
        op_intent_refs = scd_client.query_operational_intent_references(resources.utm_client, vol4)
    except (ValueError, scd_client.OperationError) as e:
        msg = 'Error querying operational intents: {}'.format(e)
        return msg, 412

    # Try to delete every operational intent found
    dss_deletion_results = {}
    deleted = set()
    for op_intent_ref in op_intent_refs:
        try:
            scd_client.delete_operational_intent_reference(resources.utm_client, op_intent_ref.id, op_intent_ref.ovn)
            dss_deletion_results[op_intent_ref.id] = 'Deleted from DSS'
            deleted.add(op_intent_ref.id)
        except scd_client.OperationError as e:
            dss_deletion_results[op_intent_ref.id] = str(e)

    # Delete corresponding flight injections and cached operational intents
    with db as tx:
        flights_to_delete = []
        for flight_id, record in tx.flights.items():
            if record.op_intent_reference.id in deleted:
                flights_to_delete.append(flight_id)
        for flight_id in flights_to_delete:
            del tx.flights[flight_id]

        cache_deletions = []
        for op_intent_id in deleted:
            if op_intent_id in tx.cached_operations:
                del tx.cached_operations[op_intent_id]
                cache_deletions.append(op_intent_id)

    return flask.jsonify({
        'dss_deletions': dss_deletion_results,
        'flight_deletions': flights_to_delete,
        'cache_deletions': cache_deletions,
    }), 200


from . import routes_scdsc
from . import routes_injection
