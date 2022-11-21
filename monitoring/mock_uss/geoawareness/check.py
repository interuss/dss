import logging
from typing import List, Dict
from uas_standards.interuss.automated_testing.geo_awareness.v1.api import (
    GeozonesCheckResultGeozone,
    GeozonesCheckRequest,
    GeozoneHttpsSourceFormat,
    GeozoneSourceResponseResult,
)
from monitoring.mock_uss.geoawareness.ed269 import evaluate_source
from monitoring.mock_uss.geoawareness.database import db, SourceRecord, Database


logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


def combine_results(
    r1: GeozonesCheckResultGeozone, r2: GeozonesCheckResultGeozone
) -> GeozonesCheckResultGeozone:
    """
    Logical OR results combination.
    Present == True and Absent == False
    """
    if GeozonesCheckResultGeozone.Present in [r1, r2]:
        return GeozonesCheckResultGeozone.Present
    return GeozonesCheckResultGeozone.Absent


def check_geozones(req: GeozonesCheckRequest) -> List[GeozonesCheckResultGeozone]:
    sources: Dict[str, SourceRecord] = Database.get_sources(db)

    results: List[GeozonesCheckResultGeozone] = [
        GeozonesCheckResultGeozone.Absent
    ] * len(req.checks)

    for i, check in enumerate(req.checks):
        logger.info(f"  Evaluating check {i}: {check}")

        result = GeozonesCheckResultGeozone.Absent
        for j, (source_id, source) in enumerate(sources.items()):
            if source.state != GeozoneSourceResponseResult.Ready:
                logger.debug(
                    f" {j+1}. Source {source_id} is not ready ({source.state}). Skip."
                )
                continue

            fmt = source.definition.https_source.format
            if fmt == GeozoneHttpsSourceFormat.ED_269:
                logger.debug(f" {j+1}. ED269 source {source_id} ready.")
                result = combine_results(
                    result, evaluate_source(source, check.filterSets)
                )
            else:
                logger.debug(
                    f" {j+1}. Source {source_id} not in supported format {fmt}. Skip."
                )

        results[i] = result

    if len(req.checks) != len(results):
        raise ValueError(
            f"Number of checks ({len(req.checks)}) do not match number of results ({len(results)})"
        )
    logger.debug(f"results: {list(map(str, results))}")
    return results
