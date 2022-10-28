import flask
import requests
from uas_standards.eurocae_ed269 import ED269Schema
from uas_standards.interuss.automated_testing.geo_awareness.v1.api import (
    GeozoneSourceResponseResult,
    CreateGeozoneSourceRequest,
    GeozoneHttpsSourceFormat,
    GeozoneSourceResponse,
)
from monitoring.mock_uss.geoawareness.database import (
    db,
    ExistingRecordException,
    Database,
)


def get_geozone_source(geozone_source_id: str):
    """This handler returns the state of a geozone source"""

    source = Database.get_source(db, geozone_source_id)
    if source is None:
        return f"source {geozone_source_id} not found or deleted", 404
    return (
        GeozoneSourceResponse(result=GeozoneSourceResponseResult.Ready),
        200,
    )


def create_geozone_source(id, source_definition: CreateGeozoneSourceRequest):
    """This handler creates and activates a geozone source"""

    try:
        source = Database.insert_source(
            db, id, source_definition, GeozoneSourceResponseResult.Activating
        )
    except ExistingRecordException:
        return f"source {id} already exists in database", 409

    if "https_source" in source.definition:
        try:
            raw_data = requests.get(source.definition.https_source.url).json()
            if source.definition.https_source.format == GeozoneHttpsSourceFormat.ED_269:
                geozones = ED269Schema.from_dict(raw_data)
                Database.update_source_geozone_ed269(db, id, geozones)
                source = Database.update_source_state(
                    db, id, GeozoneSourceResponseResult.Ready
                )
        except ValueError as e:
            source = Database.update_source_state(
                db,
                id,
                GeozoneSourceResponseResult.Error,
                f"Unable to download and parse {source.definition.https_source.url}: {str(e)}",
            )

    else:
        source = Database.update_source_state(
            db,
            id,
            GeozoneSourceResponseResult.Error,
            f"Unsupported source definition. https_source only",
        )
        return GeozoneSourceResponse(result=source.state, message=source.message), 400

    return GeozoneSourceResponse(
        result=source.state, message=source.get("message", None)
    )


def delete_geozone_source(geozone_source_id):
    """This handler deactivates and deletes a geozone source"""

    deleted_id = Database.delete_source(db, geozone_source_id)

    if deleted_id is None:
        return f"source {geozone_source_id} not found", 404

    return (
        GeozoneSourceResponse(result=GeozoneSourceResponseResult.Deactivating),
        200,
    )
