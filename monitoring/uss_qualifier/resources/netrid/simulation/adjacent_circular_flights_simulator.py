import random
from datetime import timedelta
from typing import List

import arrow
from pyproj import Geod, Transformer, Proj
import shapely.geometry
from shapely.geometry import Point, Polygon

from implicitdict import ImplicitDict
from monitoring.uss_qualifier.resources.netrid.flight_data import (
    FullFlightRecord,
    FlightRecordCollection,
    AdjacentCircularFlightsSimulatorConfiguration,
)
from monitoring.uss_qualifier.resources.netrid.simulation import (
    operator_flight_details,
)
from uas_standards.astm.f3411.v19.api import (
    HorizontalAccuracy,
    VerticalAccuracy,
    RIDAircraftPosition,
    RIDAircraftState,
    RIDHeight,
    RIDFlightDetails,
)
from .utils import (
    QueryBoundingBox,
    FlightPoint,
    GridCellFlight,
)


class AdjacentCircularFlightsSimulator:

    """A class to generate Flight Paths given a bounding box, this is the main module to generate flight path datasets, the data is generated as latitude / longitude pairs with assoiated with the flights. Additional flight metadata e.g. flight id, altitude, registration number can also be generated"""

    def __init__(self, config: AdjacentCircularFlightsSimulatorConfiguration) -> None:
        """Create an AdjacentCircularFlightsSimulator with the specified bounding box.

        Once these extents are specified, a grid will be created with two rows.  The idea is that multiple flights tracks will be created within the extents.

        Raises:
        ValueError: If bounding box has more area than a 500m x 500m square.
        """
        self.reference_time = arrow.get(config.reference_time.datetime)
        if config.random_seed is None:
            self.random = random
        else:
            self.random = random.Random(x=config.random_seed)
        self.minx = config.minx
        self.miny = config.miny
        self.maxx = config.maxx
        self.maxy = config.maxy
        self.utm_zone = config.utm_zone

        self.altitude_agl = 50.0

        self.grid_cells_flight_tracks: List[GridCellFlight] = []

        # This object holds the name and the polygon object of the query boxes. The number of bboxes are controlled by the `box_diagonals` variable
        self.query_bboxes: List[QueryBoundingBox] = []

        self.flights: List[FullFlightRecord] = []
        self.bbox_center: List[shapely.geometry.Point] = []

        self.geod = Geod(ellps="WGS84")

        self.input_extents_valid()

    def input_extents_valid(self) -> None:
        """This method checks if the input extents are valid i.e. small enough, if the extent is too large, we reject them, at the moment it checks for extents less than 500m x 500m square but can be changed as necessary."""

        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        area = abs(self.geod.geometry_area_perimeter(box)[0])

        # Have a area less than 500m x 500m square and more than 300m x 300m square to ensure a 50 m diameter tracks
        if (area) < 250000 and (area) > 90000:
            return
        else:
            raise ValueError(
                "The extents provided are not of the correct size, please provide extents that are less than 500m x 500m and more than 300m x 300m square"
            )

    def generate_query_bboxes(self):
        """For the differnet Remote ID checks: No, we need to generate three bounding boxes for the display provider, this method generates the 1 km diagonal length bounding box"""
        # Get center of of the bounding box that is inputted into the generator
        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        center = box.centroid
        self.bbox_center.append(center)
        # Transform to geographic co-ordinates to get areas
        transformer = Transformer.from_crs("epsg:4326", "epsg:3857")
        transformed_x, transformed_y = transformer.transform(center.x, center.y)
        pt = Point(transformed_x, transformed_y)
        # Now we have a point, we can buffer the point and create bounding boxes of the buffer to get the appropriate polygons, more than three boxes can be created, for the tests three will suffice.
        now = self.reference_time

        box_diagonals = [
            {
                "length": 150,
                "name": "zoomed_in_detail",
                "timestamp_after": now + timedelta(seconds=60),
                "timestamp_before": now + timedelta(seconds=90),
            },
            {
                "length": 380,
                "name": "whole_flight_area",
                "timestamp_after": now + timedelta(seconds=30),
                "timestamp_before": now + timedelta(seconds=60),
            },
            {
                "length": 3000,
                "name": "too_large_query",
                "timestamp_after": now + timedelta(seconds=10),
                "timestamp_before": now + timedelta(seconds=30),
            },
        ]

        for box_id, box_diagonal in enumerate(box_diagonals):
            # Buffer the point with the appropriate length
            buffer = pt.buffer(box_diagonal["length"])
            buffer_bounds = buffer.bounds
            buffer_bounds_polygon = shapely.geometry.box(
                buffer_bounds[0], buffer_bounds[1], buffer_bounds[2], buffer_bounds[3]
            )
            buffer_points = zip(
                buffer_bounds_polygon.exterior.coords.xy[0],
                buffer_bounds_polygon.exterior.coords.xy[1],
            )
            proj_buffer_points = []
            # reproject back to ESPG 4326
            transformer2 = Transformer.from_crs("epsg:3857", "epsg:4326")
            for point in buffer_points:
                x = point[0]
                y = point[1]
                x, y = transformer2.transform(x, y)
                proj_buffer_points.append((x, y))

            buffered_box = Polygon(proj_buffer_points)
            # Get the bounds of the buffered box, this is the one that will will be fed to the remote ID display provider to query
            self.query_bboxes.append(
                QueryBoundingBox(
                    name=box_diagonals[box_id]["name"],
                    shape=buffered_box,
                    timestamp_after=box_diagonals[box_id]["timestamp_after"].datetime,
                    timestamp_before=box_diagonals[box_id]["timestamp_before"].datetime,
                )
            )

    def generate_flight_speed_bearing(
        self, adjacent_points: List, delta_time_secs: int
    ) -> List[float]:
        """A method to generate flight speed, assume that the flight has to traverse two adjecent points in x number of seconds provided, calculating speed in meters / second. It also generates bearing between this and next point, this is used to populate the 'track' paramater in the Aircraft State JSON."""

        first_point = adjacent_points[0]
        second_point = adjacent_points[1]

        fwd_azimuth, back_azimuth, adjacent_point_distance_mts = self.geod.inv(
            first_point.x, first_point.y, second_point.x, second_point.y
        )

        speed_mts_per_sec = adjacent_point_distance_mts / delta_time_secs
        speed_mts_per_sec = float("{:.2f}".format(speed_mts_per_sec))

        if fwd_azimuth < 0:
            fwd_azimuth = 360 + fwd_azimuth

        return [speed_mts_per_sec, fwd_azimuth]

    def utm_converter(
        self, shapely_shape: shapely.geometry, inverse: bool = False
    ) -> shapely.geometry.shape:
        """A helper function to convert from lat / lon to UTM coordinates for buffering. tracks. This is the UTM projection (https://en.wikipedia.org/wiki/Universal_Transverse_Mercator_coordinate_system), we use Zone 33T which encompasses Switzerland, this zone has to be set for each locale / city. Adapted from https://gis.stackexchange.com/questions/325926/buffering-geometry-with-points-in-wgs84-using-shapely"""

        proj = Proj(proj="utm", zone=self.utm_zone, ellps="WGS84", datum="WGS84")

        geo_interface = shapely_shape.__geo_interface__
        point_or_polygon = geo_interface["type"]
        coordinates = geo_interface["coordinates"]
        if point_or_polygon == "Polygon":
            new_coordinates = [
                [proj(*point, inverse=inverse) for point in linring]
                for linring in coordinates
            ]
        elif point_or_polygon == "Point":
            new_coordinates = proj(*coordinates, inverse=inverse)
        else:
            raise RuntimeError(
                "Unexpected geo_interface type: {}".format(point_or_polygon)
            )

        return shapely.geometry.shape(
            {"type": point_or_polygon, "coordinates": tuple(new_coordinates)}
        )

    def generate_flight_grid_and_path_points(
        self, altitude_of_ground_level_wgs_84: float
    ):
        """Generate a series of boxes (grid) within the given bounding box to have areas for different flight tracks within each box"""
        # Compute the box where the flights will be created. For a the sample bounds given, over Bern, Switzerland, a division by 2 produces a cell_size of 0.0025212764739985793, a division of 3 is 0.0016808509826657196 and division by 4 0.0012606382369992897. As the cell size goes smaller more number of flights can be accomodated within the grid. For the study area bounds we build a 3x2 box for six flights by creating 3 column 2 row grid.
        N_COLS = 3
        N_ROWS = 2
        cell_size_x = (self.maxx - self.minx) / (N_COLS)  # create three columns
        cell_size_y = (self.maxy - self.miny) / (N_ROWS)  # create two rows
        grid_cells = []
        for u0 in range(0, N_COLS):  # 3 columns
            x0 = self.minx + (u0 * cell_size_x)
            for v0 in range(0, N_ROWS):  # 2 rows
                y0 = self.miny + (v0 * cell_size_y)
                x1 = x0 + cell_size_x
                y1 = y0 + cell_size_y
                grid_cells.append(shapely.geometry.box(x0, y0, x1, y1))

        all_grid_cell_tracks = []
        """ For each of the boxes (grid) allocated to the operator, get the centroid and buffer to generate a flight path. A 50 m radius is provided to have flight paths within each of the boxes """
        # Iterate over the flight_grid
        for grid_cell in grid_cells:
            center = grid_cell.centroid
            center_utm = self.utm_converter(center)
            buffer_shape_utm = center_utm.buffer(50)
            buffered_path = self.utm_converter(buffer_shape_utm, inverse=True)
            altitude = (
                altitude_of_ground_level_wgs_84 + self.altitude_agl
            )  # meters WGS 84
            flight_points_with_altitude = []
            x, y = buffered_path.exterior.coords.xy

            for coord in range(0, len(x)):
                cur_coord = coord
                next_coord = coord + 1
                next_coord = 0 if next_coord == len(x) else next_coord
                adjacent_points = [
                    Point(x[cur_coord], y[cur_coord]),
                    Point(x[next_coord], y[next_coord]),
                ]
                flight_speed, bearing = self.generate_flight_speed_bearing(
                    adjacent_points=adjacent_points, delta_time_secs=1
                )

                flight_points_with_altitude.append(
                    FlightPoint(
                        lat=y[coord],
                        lng=x[coord],
                        alt=altitude,
                        speed=flight_speed,
                        bearing=bearing,
                    )
                )

            all_grid_cell_tracks.append(
                GridCellFlight(bounds=grid_cell, track=flight_points_with_altitude)
            )

        self.grid_cells_flight_tracks = all_grid_cell_tracks

    def generate_flight_details(self, id: str) -> RIDFlightDetails:
        """This class generates details of flights and operator details for a flight, this data is required for identifying flight, operator and operation"""

        my_flight_details_generator = (
            operator_flight_details.OperatorFlightDataGenerator(self.random)
        )

        # TODO: Put operator_location in center of circle rather than stacking operators of all flights on top of each other
        return RIDFlightDetails(
            id=id,
            serial_number=my_flight_details_generator.generate_serial_number(),
            operation_description=my_flight_details_generator.generate_operation_description(),
            operator_location=my_flight_details_generator.generate_operator_location(
                centroid=self.bbox_center[0]
            ),
            operator_id=my_flight_details_generator.generate_operator_id(),
            registration_number=my_flight_details_generator.generate_registration_number(),
        )

    def generate_rid_state(self, duration):
        """

        This method generates rid_state objects that can be submitted as flight telemetry


        """
        all_flight_telemetry: List[List[RIDAircraftState]] = []
        flight_track_details = {}  # Develop a index of flight length and their index
        # Store where on the track the current index is, since the tracks are circular, once the end of the track is reached, the index is reset to 0 to indicate beginning of the track again.
        flight_current_index = {}
        # Get the number of flights
        num_flights = len(self.grid_cells_flight_tracks)
        time_increment_seconds = 1  # the number of seconds it takes to go from one point to next on the track
        now = self.reference_time
        now_isoformat = now.isoformat()
        for i in range(num_flights):
            flight_positions_len = len(self.grid_cells_flight_tracks[i].track)

            # in a circular flight pattern increment direction
            angle_increment = 360 / flight_positions_len

            # the resolution of track is 1 degree minimum
            angle_increment = 1.0 if angle_increment == 0.0 else angle_increment

            if i not in flight_track_details:
                flight_track_details[i] = {}
            flight_track_details[i]["track_length"] = flight_positions_len
            flight_current_index[i] = 0
            all_flight_telemetry.append([])

        timestamp = now
        for j in range(duration):
            timestamp = timestamp.shift(seconds=time_increment_seconds)

            timestamp_isoformat = timestamp.isoformat()

            for k in range(num_flights):
                list_end = (
                    flight_track_details[k]["track_length"] - flight_current_index[k]
                )

                if list_end != 1:
                    flight_point = self.grid_cells_flight_tracks[k].track[
                        flight_current_index[k]
                    ]
                    aircraft_position = RIDAircraftPosition(
                        lat=flight_point.lat,
                        lng=flight_point.lng,
                        alt=flight_point.alt,
                        accuracy_h=HorizontalAccuracy.HAUnknown,
                        accuracy_v=VerticalAccuracy.VAUnknown,
                        extrapolated=False,
                    )
                    aircraft_height = RIDHeight(
                        distance=self.altitude_agl, reference="TakeoffLocation"
                    )

                    rid_aircraft_state = RIDAircraftState(
                        timestamp=timestamp_isoformat,
                        operational_status="Airborne",
                        position=aircraft_position,
                        height=aircraft_height,
                        track=flight_point.bearing,
                        speed=flight_point.speed,
                        timestamp_accuracy=0.0,
                        speed_accuracy="SA3mps",
                        vertical_speed=0.0,
                    )

                    all_flight_telemetry[k].append(rid_aircraft_state)

                    flight_current_index[k] += 1
                else:
                    flight_current_index[k] = 0

        flights = []
        for m in range(num_flights):
            flight = FullFlightRecord(
                reference_time=now_isoformat,
                states=all_flight_telemetry[m],
                flight_details=self.generate_flight_details(id=str(m)),
                aircraft_type="Helicopter",
            )
            flights.append(flight)

        self.flights = flights


def generate_aircraft_states(
    config: AdjacentCircularFlightsSimulatorConfiguration,
) -> FlightRecordCollection:
    my_path_generator = AdjacentCircularFlightsSimulator(config)

    my_path_generator.generate_flight_grid_and_path_points(
        altitude_of_ground_level_wgs_84=config.altitude_of_ground_level_wgs_84
    )
    my_path_generator.generate_query_bboxes()

    my_path_generator.generate_rid_state(duration=30)
    flights = my_path_generator.flights

    result = FlightRecordCollection(flights=flights)

    # Fix type errors (TODO: fix these at the source rather than here)
    return ImplicitDict.parse(result, FlightRecordCollection)
