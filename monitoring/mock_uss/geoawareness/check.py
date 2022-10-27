import logging
from typing import List, Dict
from monitoring.mock_uss.geoawareness.ed269 import evaluate_source

from monitoring.mock_uss.geoawareness.database import db, SourceRecord, Database
from monitoring.monitorlib.geoawareness_automated_testing.api import (
    GeozonesCheckRequest,
    GeozoneSourceState,
    HttpsSourceFormat,
    GeozonesCheckResultName,
)

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


FEET_PER_METER = 1 / 0.3048


def combine_results(
    r1: GeozonesCheckResultName, r2: GeozonesCheckResultName
) -> GeozonesCheckResultName:
    """
    Logical OR results combination.
    Present == True and Absent == False
    """
    if GeozonesCheckResultName.Present in [r1, r2]:
        return GeozonesCheckResultName.Present
    return GeozonesCheckResultName.Absent


def check_geozones(req: GeozonesCheckRequest) -> List[GeozonesCheckResultName]:
    sources: Dict[str, SourceRecord] = Database.get_sources(db)

    results: List[GeozonesCheckResultName] = [GeozonesCheckResultName.Absent] * len(
        req.checks
    )

    for i, check in enumerate(req.checks):
        logger.info(f"  Evaluating check {i}: {check}")

        result = GeozonesCheckResultName.Absent
        for j, (source_id, source) in enumerate(sources.items()):
            if source.state != GeozoneSourceState.Ready:
                logger.debug(
                    f" {j+1}. Source {source_id} is not ready ({source.state}). Skip."
                )
                continue

            fmt = source.definition.https_source.format
            if fmt == HttpsSourceFormat.Ed269:
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
