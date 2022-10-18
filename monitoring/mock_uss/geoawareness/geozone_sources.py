import flask
import requests

from monitoring.mock_uss.geoawareness.database import (
    db,
    ExistingRecordException,
    Database,
)
from monitoring.mock_uss.geoawareness.parsers.ed269 import ED269Schema
from monitoring.monitorlib.geoawareness_automated_testing import api as geoawareness_api
from monitoring.monitorlib.geoawareness_automated_testing.api import (
    GeozoneSourceResponse,
    GeozoneSourceState,
    HttpsSourceFormat,
)


def get_geozone_source(geozone_source_id: str):
    """This handler returns the state of a geozone source"""

    source = Database.get_source(db, geozone_source_id)
    if source is None:
        return f"source {geozone_source_id} not found or deleted", 404
    return (
        flask.jsonify(GeozoneSourceResponse({"result": GeozoneSourceState.Ready})),
        200,
    )


def create_geozone_source(
    id, source_definition: geoawareness_api.GeozoneSourceDefinition
):
    """This handler creates and activates a geozone source"""

    try:
        source = Database.insert_source(
            db, id, source_definition, GeozoneSourceState.Activating
        )
    except ExistingRecordException:
        return f"source {id} already exists in database", 409

    if "https_source" in source.definition:
        try:
            raw_data = requests.get(source.definition.https_source.url).json()
            if source.definition.https_source.format == HttpsSourceFormat.Ed269:
                geozones = ED269Schema.from_dict(raw_data)
                Database.update_source_geozone_ed269(db, id, geozones)
                source = Database.update_source_state(db, id, GeozoneSourceState.Ready)
        except ValueError as e:
            source = Database.update_source_state(
                db,
                id,
                GeozoneSourceState.Error,
                f"Unable to download and parse {source.definition.https_source.url}: {str(e)}",
            )

    else:
        source = Database.update_source_state(
            db,
            id,
            GeozoneSourceState.Error,
            f"Unsupported source definition. https_source only",
        )

    return GeozoneSourceResponse(
        result=source.state, message=source.get("message", None)
    )


def delete_geozone_source(geozone_source_id):
    """This handler deactivates and deletes a geozone source"""

    deleted_id = Database.delete_source(db, geozone_source_id)

    if deleted_id is None:
        return f"source {geozone_source_id} not found", 404

    return (
        flask.jsonify(
            GeozoneSourceResponse({"result": GeozoneSourceState.Deactivating})
        ),
        200,
    )
