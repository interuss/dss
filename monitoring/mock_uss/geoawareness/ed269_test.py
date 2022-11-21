import pytest
from s2sphere import LatLng
from uas_standards.interuss.automated_testing.geo_awareness.v1.api import (
    Position,
    ED269Filters,
)
from implicitdict import StringBasedDateTime
from monitoring.mock_uss.geoawareness.ed269 import (
    evaluate_non_spacetime,
    evaluate_position,
    convert_distance,
    evaluate_timing,
)
from monitoring.monitorlib.geo import unflatten
from uas_standards.eurocae_ed269 import (
    UASZoneVersion,
    UomDimensions,
    VerticalReferenceType,
    CircleOrPolygonType,
    UASZoneAirspaceVolume,
    ApplicableTimePeriod,
    YESNO,
)


def test_convert_units():
    assert convert_distance(1, UomDimensions.M, UomDimensions.M) == 1
    assert convert_distance(1, UomDimensions.M, UomDimensions.FT) == pytest.approx(
        3.28084
    )
    assert convert_distance(1, UomDimensions.FT, UomDimensions.M) == pytest.approx(
        0.3048
    )
    assert convert_distance(1, UomDimensions.FT, UomDimensions.FT) == 1


# Airspace Volumes
circle1 = UASZoneAirspaceVolume(
    uomDimensions=UomDimensions.M,
    lowerLimit=0,
    lowerVerticalReference=VerticalReferenceType.AGL,
    upperLimit=200,
    upperVerticalReference=VerticalReferenceType.AGL,
    horizontalProjection=CircleOrPolygonType(
        type="Circle", center=[6.143158, 46.204391], radius=3000  # lng / lat
    ),
)

polygon1 = UASZoneAirspaceVolume(
    uomDimensions=UomDimensions.M,
    lowerLimit=0,
    lowerVerticalReference=VerticalReferenceType.AGL,
    upperLimit=200,
    upperVerticalReference=VerticalReferenceType.AGL,
    horizontalProjection=CircleOrPolygonType(
        type="Polygon",
        coordinates=[
            [  # lng / lat
                [
                    6.14919923049257,
                    46.18862418135734,
                ],
                [
                    6.141361756910044,
                    46.17415353875836,
                ],
                [
                    6.1609554408669,
                    46.1750580655015,
                ],
                [
                    6.180549124822704,
                    46.194049690428386,
                ],
                [
                    6.177936633628519,
                    46.213034754532885,
                ],
                [
                    6.143974248104229,
                    46.209419057660796,
                ],
                [
                    6.104786880192535,
                    46.209419057660796,
                ],
                [
                    6.14919923049257,
                    46.18862418135734,
                ],
            ],
        ],
    ),
)


def test_evaluate_position_circle():
    other_fields = {
        "identifier": "I",
        "country": "CHE",
        "type": "COMMON",
        "zoneAuthority": [],
        "applicability": [],
        "restriction": "PROHIBITED",
    }

    # TODO: Test height AMSL
    # TODO: Test height AGL
    # TODO: Test default lowerlimit 0.
    # TODO: Test default upperlimit is unbound.

    # Point inside the circle
    assert (
        evaluate_position(
            UASZoneVersion(geometry=[circle1], **other_fields),
            Position(
                uomDimensions=UomDimensions.M,
                verticalReferenceType=VerticalReferenceType.AGL,
                height=100,
                longitude=6.143158,
                latitude=46.204391,
            ),
        )
        is True
    )

    # Point outside the circle
    p = unflatten(LatLng.from_degrees(46.204391, 6.143158), (3000, 3000))
    assert (
        evaluate_position(
            UASZoneVersion(geometry=[circle1], **other_fields),
            Position(
                uomDimensions=UomDimensions.M,
                verticalReferenceType=VerticalReferenceType.AGL,
                height=100,
                longitude=p.lng().degrees,
                latitude=p.lat().degrees,
            ),
        )
        is False
    )


def test_evaluate_position_polygon():
    other_fields = {
        "identifier": "I",
        "country": "CHE",
        "type": "COMMON",
        "zoneAuthority": [],
        "applicability": [],
        "restriction": "PROHIBITED",
    }

    # Point inside the polygon
    p = LatLng.from_degrees(46.204391, 6.143158)
    assert (
        evaluate_position(
            UASZoneVersion(geometry=[polygon1], **other_fields),
            Position(
                uomDimensions=UomDimensions.M,
                verticalReferenceType=VerticalReferenceType.AGL,
                height=100,
                longitude=p.lng().degrees,
                latitude=p.lat().degrees,
            ),
        )
        is True
    )

    # Point outside the polygon
    p = LatLng.from_degrees(46.188236800723985, 6.143868308377478)
    assert (
        evaluate_position(
            UASZoneVersion(geometry=[polygon1], **other_fields),
            Position(
                uomDimensions=UomDimensions.M,
                verticalReferenceType=VerticalReferenceType.AGL,
                height=100,
                longitude=p.lng().degrees,
                latitude=p.lat().degrees,
            ),
        )
        is False
    )


def test_evaluate_position_multigeometry():
    other_fields = {
        "identifier": "I",
        "country": "CHE",
        "type": "COMMON",
        "zoneAuthority": [],
        "applicability": [],
        "restriction": "PROHIBITED",
    }

    # Point inside the polygon
    p = LatLng.from_degrees(46.204391, 6.143158)
    assert (
        evaluate_position(
            UASZoneVersion(geometry=[polygon1, circle1], **other_fields),
            Position(
                uomDimensions=UomDimensions.M,
                verticalReferenceType=VerticalReferenceType.AGL,
                height=100,
                longitude=p.lng().degrees,
                latitude=p.lat().degrees,
            ),
        )
        is True
    )

    # Point inside the circle
    p = LatLng.from_degrees(46.204391, 6.143158)
    assert (
        evaluate_position(
            UASZoneVersion(geometry=[polygon1, circle1], **other_fields),
            Position(
                uomDimensions=UomDimensions.M,
                verticalReferenceType=VerticalReferenceType.AGL,
                height=100,
                longitude=p.lng().degrees,
                latitude=p.lat().degrees,
            ),
        )
        is True
    )

    # Point outside the polygon and outside the circle
    p = LatLng.from_degrees(46.38236800723985, 6.453868308377478)
    assert (
        evaluate_position(
            UASZoneVersion(geometry=[polygon1, circle1], **other_fields),
            Position(
                uomDimensions=UomDimensions.M,
                verticalReferenceType=VerticalReferenceType.AGL,
                height=100,
                longitude=p.lng().degrees,
                latitude=p.lat().degrees,
            ),
        )
        is False
    )


def test_evaluate_timing():
    other_fields = {
        "country": "CHE",
        "type": "COMMON",
        "geometry": [],
        "zoneAuthority": [],
        "restriction": "PROHIBITED",
    }

    # Permanent
    assert (
        evaluate_timing(
            feature=UASZoneVersion(
                applicability=[ApplicableTimePeriod(permanent=YESNO.YES)],
                identifier="Permanent",
                **other_fields,
            ),
            after=None,
            before=None,
        )
        is True
    )

    # Temporary applicability
    d1 = StringBasedDateTime("2022-02-01T00:00:00Z")
    d2 = StringBasedDateTime("2022-02-02T00:00:00Z")
    d3 = StringBasedDateTime("2022-02-03T00:00:00Z")
    d4 = StringBasedDateTime("2022-02-04T00:00:00Z")

    test_ranges = [
        # Tuple(feature start, feature end, after filter, before filter, expected outcome)
        # before filter only - in scope
        (d1, d2, None, d3, True),
        # before filter only - out of scope
        (d2, d3, None, d1, False),
        # after filter only - in scope
        (d2, d3, d1, None, True),
        # after filter only - out of scope
        (d1, d2, d3, None, False),
        # after and before filters - within scope
        (d2, d3, d1, d4, True),
        # after and before filters - encompass scope (start before and end after)
        (d1, d4, d2, d3, True),
        # after and before filters - partially in scope (end after)
        (d2, d4, d1, d3, True),
        # after and before filters - out of scope (before)
        (d1, d2, d3, d4, False),
        # after and before filters - out of scope (after)
        (d3, d4, d1, d2, False),
        # no filter - present
        (d3, d4, None, None, True),
    ]

    for i, t in enumerate(test_ranges):
        assert (
            evaluate_timing(
                feature=UASZoneVersion(
                    identifier=f"TMP{i}",
                    applicability=[
                        ApplicableTimePeriod(
                            permanent=YESNO.NO, startDateTime=t[0], endDateTime=t[1]
                        )
                    ],
                    **other_fields,
                ),
                after=t[2],
                before=t[3],
            )
            is t[4]
        )

    # Multiple applicability
    assert (
        evaluate_timing(
            feature=UASZoneVersion(
                applicability=[
                    ApplicableTimePeriod(
                        permanent=YESNO.NO, startDateTime=d2, endDateTime=d3
                    ),  # No match
                    ApplicableTimePeriod(
                        permanent=YESNO.NO, startDateTime=d1, endDateTime=d3
                    ),  # Match
                ],
                identifier="Permanent",
                **other_fields,
            ),
            after=None,
            before=d2,
        )
        is True
    )


def test_evaluate_non_spacetime():
    other_fields = {
        "identifier": "I",
        "country": "CHE",
        "type": "COMMON",
        "zoneAuthority": [],
        "applicability": [],
        "geometry": [],
    }

    # uSpaceClass match
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(uSpaceClass="C1", restriction="PROHIBITED", **other_fields),
            ED269Filters(uSpaceClass="C1"),
        )
        is True
    )

    # uSpaceClass mismatch
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(uSpaceClass="C1", restriction="PROHIBITED", **other_fields),
            ED269Filters(uSpaceClass="C2"),
        )
        is False
    )

    # uSpaceClass as array
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(
                uSpaceClass=["C1", "C2"], restriction="PROHIBITED", **other_fields
            ),
            ED269Filters(uSpaceClass="C2"),
        )
        is True
    )

    # uSpaceClass as JSON array in a string
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(
                uSpaceClass='["C1", "C2"]', restriction="PROHIBITED", **other_fields
            ),
            ED269Filters(uSpaceClass="C2"),
        )
        is True
    )

    # uSpaceClass as Python array in a string
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(
                uSpaceClass="['C1', 'C2']", restriction="PROHIBITED", **other_fields
            ),
            ED269Filters(uSpaceClass="C2"),
        )
        is True
    )

    # No filter
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(uSpaceClass="C1", restriction="PROHIBITED", **other_fields),
            ED269Filters(),
        )
        is True
    )

    # No filter and no uSpaceClass
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(restriction="PROHIBITED", **other_fields), ED269Filters()
        )
        is True
    )

    # Missing value
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(restriction="PROHIBITED", **other_fields),
            ED269Filters(uSpaceClass="C1"),
        )
        is False
    )

    # Single match with single acceptable restriction
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(uSpaceClass="C1", restriction="PROHIBITED", **other_fields),
            ED269Filters(acceptableRestrictions=["PROHIBITED"]),
        )
        is True
    )

    # Single match in multiple acceptable restrictions
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(uSpaceClass="C1", restriction="PROHIBITED", **other_fields),
            ED269Filters(acceptableRestrictions=["REQ_AUTHORISATION", "PROHIBITED"]),
        )
        is True
    )

    # Unacceptable restriction only
    assert (
        evaluate_non_spacetime(
            UASZoneVersion(uSpaceClass="C1", restriction="PROHIBITED", **other_fields),
            ED269Filters(acceptableRestrictions=["REQ_AUTHORISATION"]),
        )
        is False
    )
