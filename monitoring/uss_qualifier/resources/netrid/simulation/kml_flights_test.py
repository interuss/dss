"""Unit tests to test create_flight_record_from_kml module using pytest.

Testing can be invoked from the command line using:
`pytest [test_*|*_test.py file/filepath]`
"""

from . import kml_flights as frk

PACKAGE = "monitoring.uss_qualifier.resources.netrid.simulation"


def test_get_interpolated_value(mocker):
    mocker.patch(
        f"{PACKAGE}.kml_flights.get_polygons_distances_from_point",
        return_value=[1.1, 2.0, 3.3, 0.0],
    )
    result_alt = frk.get_interpolated_value("point", "polygons", [10, 20, 30, 40])
    assert result_alt == 40

    mocker.patch(
        f"{PACKAGE}.kml_flights.get_polygons_distances_from_point",
        return_value=[10, 50, 100],
    )
    result_alt = frk.get_interpolated_value("point", "polygons", [10, 20, 30])
    assert abs(result_alt - 13.1) < 0.1

    mocker.patch(
        f"{PACKAGE}.kml_flights.get_polygons_distances_from_point",
        return_value=[10, 50, 100],
    )
    result_alt = frk.get_interpolated_value("point", "polygons", [140, 125, 116])
    assert abs(result_alt - 135.84) < 0.1


def test_get_track_angle():
    # towards straight north
    point1 = (1, 1)
    point2 = (1, 3)
    assert frk.get_track_angle(point1, point2) == 0

    # towards south
    point1 = (4, 4)
    point2 = (4, -4)
    assert frk.get_track_angle(point1, point2) == 180

    # 1st coordinate, equal distance from x and y coordinates.
    point1 = (1, 1)
    point2 = (3, 3)
    assert frk.get_track_angle(point1, point2) == 45

    # 1st coordinate, north-east, more inclined towards north.
    point1 = (1, 1)
    point2 = (3, 4)
    assert (
        frk.get_track_angle(point1, point2) > 0
        and frk.get_track_angle(point1, point2) < 45
    )

    # moving towards south-east
    point1 = (1, -1)
    point2 = (2, -4)
    assert (
        frk.get_track_angle(point1, point2) > 90
        and frk.get_track_angle(point1, point2) < 180
    )
    point1 = (1, 4)
    point2 = (2, 1)
    assert (
        frk.get_track_angle(point1, point2) > 90
        and frk.get_track_angle(point1, point2) < 180
    )

    # moving towards south-west
    point1 = (4, 4)
    point2 = (1, 1)
    assert (
        frk.get_track_angle(point1, point2) > 180
        and frk.get_track_angle(point1, point2) < 270
    )
    point1 = (-1, -1)
    point2 = (-4, -4)
    assert (
        frk.get_track_angle(point1, point2) > 180
        and frk.get_track_angle(point1, point2) < 270
    )

    # towards north-west
    point1 = (-1, 1)
    point2 = (-4, 4)
    assert (
        frk.get_track_angle(point1, point2) > 270
        and frk.get_track_angle(point1, point2) < 360
    )
    point1 = (4, -4)
    point2 = (1, -1)
    assert (
        frk.get_track_angle(point1, point2) > 270
        and frk.get_track_angle(point1, point2) < 360
    )
