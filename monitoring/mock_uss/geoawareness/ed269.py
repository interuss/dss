import ast
import json
import logging
from datetime import datetime
from typing import List, Optional, Union

import s2sphere
from s2sphere import LatLng
from shapely.geometry import Point, Polygon

from implicitdict import StringBasedDateTime
from monitoring.mock_uss.geoawareness.database import SourceRecord
from monitoring.monitorlib.geo import flatten
from uas_standards.interuss.automated_testing.geo_awareness.v1.api import (
    GeozonesFilterSet,
    Position,
    ED269Filters,
    GeozonesCheckResultGeozone,
    GeozoneSourceResponseResult,
)
from uas_standards.eurocae_ed269 import (
    UASZoneVersion,
    HorizontalProjectionType,
    UomDimensions,
    YESNO,
)

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


FEET_PER_METER = 1 / 0.3048


def convert_distance(
    distance: float,
    uom_dimensions: UomDimensions,
    desired_dimensions: UomDimensions = UomDimensions.M,
):
    if uom_dimensions == desired_dimensions:
        return distance
    if uom_dimensions == UomDimensions.M:
        return distance * FEET_PER_METER
    if uom_dimensions == UomDimensions.FT:
        return distance / FEET_PER_METER
    raise ValueError(
        f"Unsupported UomDimensions: Received {uom_dimensions} - Expected: {desired_dimensions}"
    )


def evaluate_position(feature: UASZoneVersion, position: Optional[Position]) -> bool:
    logger.debug(f"      _evaluate_position: position:{position} feature:{feature}")
    if position is None:
        logger.debug(f"        * position is None => True")
        return True

    # TODO: Implement height check in AGL and AMSL

    for g in feature.geometry:
        if g.horizontalProjection.type == HorizontalProjectionType.Circle:
            center = g.horizontalProjection.center  # Lng / Lat
            ref = LatLng.from_degrees(center[1], center[0])
            radius = convert_distance(
                g.horizontalProjection.radius, g.uomDimensions, UomDimensions.M
            )

            circle_2d = Point(0, 0).buffer(radius)
            position_2d = Point(
                flatten(ref, LatLng.from_degrees(position.latitude, position.longitude))
            )
            if position_2d.within(circle_2d):
                logger.debug(f"        * position is within circle => True")
                return True
        else:
            for coord in g.horizontalProjection.coordinates:  # Lng / Lat
                ref = s2sphere.LatLng.from_degrees(
                    coord[0][1], coord[0][0]
                )  # TODO: Use barycenter as reference instead of first point.

                polygon_2d = Polygon(
                    [
                        flatten(ref, s2sphere.LatLng.from_degrees(p[1], p[0]))
                        for p in coord  # Lng / Lat
                    ]
                )
                position_2d = Point(
                    flatten(
                        ref, LatLng.from_degrees(position.latitude, position.longitude)
                    )
                )

                if position_2d.within(polygon_2d):
                    logger.debug(f"      * position is within polygon => True")
                    return True

    logger.debug(f"      * position outside geometry => False")
    return False


def _is_in_date_range(
    start: StringBasedDateTime,
    end: StringBasedDateTime,
    after: Optional[StringBasedDateTime],
    before: Optional[StringBasedDateTime],
) -> bool:

    if after is None and before is None:
        return True

    elif after is not None and before is not None:
        if start.datetime < before.datetime and end.datetime > after.datetime:
            return True
    else:
        if after is not None and end.datetime > after.datetime:
            return True

        if before is not None and start.datetime < before.datetime:
            return True

    return False


def evaluate_timing(
    feature: UASZoneVersion,
    after: Optional[StringBasedDateTime] = None,
    before: Optional[StringBasedDateTime] = None,
) -> bool:
    logger.debug(
        f"     _evaluate_timing: after:{after} before:{before} feature:{feature}"
    )

    for a in feature.applicability:
        if a.permanent == YESNO.YES:
            logger.debug(f"     * Permanent applicability => True")
            return True

        start = a.get("startDateTime", StringBasedDateTime(datetime.min))
        end = a.get("endDateTime", StringBasedDateTime(datetime.max))

        in_range = _is_in_date_range(start, end, after, before)
        if not in_range:
            continue

        schedule = a.get("schedule", None)
        if schedule is None:
            logger.debug(f"     * Date in range without schedule => True")
            return True
        else:
            # TODO Implement schedule checks
            logger.warning(
                f"      * Date in range => {in_range} (Schedule not taken into account)"
            )
            return True
    return False


def _adjust_uspace_class(uspace_class: Union[str, List[str]]) -> List[str]:
    # TODO: Revisit when new version of ED-269 will be published.
    #  uSpaceClass field is currently defined in ED269 standard as a string.
    #  The current assumption is that uSpaceClass will be a string or some sort of an array of values.

    if isinstance(uspace_class, list):
        return uspace_class

    # Check if uSpaceClass is a JSON array in a string
    if isinstance(uspace_class, str):
        try:
            uspace_class = json.loads(uspace_class)
        except Exception:
            # Check if uSpaceClass is a Python list in as string
            try:
                # TODO: Discuss if serialization should be reviewed in implicitdict for String types to use JSON format
                uspace_class = ast.literal_eval(uspace_class)
            except Exception:
                pass

    # Ensure uSpaceClass is a list
    if isinstance(uspace_class, str):
        uspace_class = [uspace_class]

    return uspace_class


def evaluate_non_spacetime(
    feature: UASZoneVersion, ed269: Optional[ED269Filters]
) -> bool:
    """Returns True if the feature matches all provided filters"""
    logger.debug(f"     _evaluate_ed269: ed269:{ed269} feature:{feature}")

    if ed269 is None:
        return True

    uspace_class_filter = ed269.get("uSpaceClass", None)
    uspace_class: Optional[List[str]] = _adjust_uspace_class(
        feature.get("uSpaceClass", None)
    )

    if uspace_class_filter is not None:
        if uspace_class is None:
            return False
        if uspace_class is not None and uspace_class_filter not in uspace_class:
            return False

    acceptable_restrictions_filter = ed269.get("acceptableRestrictions", None)
    restriction = feature.get("restriction", None)
    if (
        acceptable_restrictions_filter is not None
        and restriction not in acceptable_restrictions_filter
    ):
        return False

    # All match or no filter
    return True


def evaluate_feature(feature: UASZoneVersion, filter_set: GeozonesFilterSet) -> bool:
    # Evaluate position
    position_match = evaluate_position(feature, filter_set.get("position", None))
    if not position_match:
        logger.debug(f"    {feature.identifier}: Position not matched - Absent")
        return False

    # Evaluate timing
    timing_match = evaluate_timing(
        feature, filter_set.get("after", None), filter_set.get("before", None)
    )
    if not timing_match:
        logger.debug(f"    {feature.identifier}: Timing not matched - Absent")
        return False

    # Evaluate non-spacetime fields
    non_spacetime_match = evaluate_non_spacetime(feature, filter_set.get("ed269", None))
    if not non_spacetime_match:
        logger.debug(f"    {feature.identifier}: ED269 not matched - Absent")
        return False

    logger.info(f"  {feature.identifier}: Present")
    return True


def evaluate_features(
    features: List[UASZoneVersion], filter_set: GeozonesFilterSet
) -> GeozonesCheckResultGeozone:
    logger.debug(f"  Evalutating {len(features)} features:")

    for i, feature in enumerate(features):
        if evaluate_feature(feature, filter_set):
            return GeozonesCheckResultGeozone.Present

    logger.info(f" => No match - Absent")
    return GeozonesCheckResultGeozone.Absent


def evaluate_source(source: SourceRecord, filter_sets: List[GeozonesFilterSet]):
    if not (
        source.state == GeozoneSourceResponseResult.Ready and "geozone_ed269" in source
    ):
        raise ValueError("Source not loaded correctly. geozone_ed269 field missing.")

    if len(filter_sets) == 0:
        return GeozonesCheckResultGeozone.Present

    features = source["geozone_ed269"]["features"]
    for f in filter_sets:
        if evaluate_features(features, f) == GeozonesCheckResultGeozone.Present:
            return GeozonesCheckResultGeozone.Present
    return GeozonesCheckResultGeozone.Absent
