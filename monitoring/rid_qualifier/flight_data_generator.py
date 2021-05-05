from shapely.geometry import Point, Polygon, LineString
import shapely.geometry
from pyproj import Geod, Transformer, Proj
import json
from pathlib import Path
from typing import List, NamedTuple, Any
import arrow
import datetime
from datetime import datetime, timedelta

from shapely.geometry.geo import shape
from monitoring.rid_qualifier.utils import QueryBoundingBox, FlightPoint, GridCellFlight
from monitoring.monitorlib.rid import AircraftPosition, AircraftHeight, AircraftState, RIDFlight
import operator_flight_details_generator as details_generator

class QueryBoundingBox(NamedTuple):
    ''' This is the object that stores details of query bounding box '''

    name: str
    shape: Polygon
    timestamp_before: timedelta
    timestamp_after: timedelta


class FlightPoint(NamedTuple):
    ''' This object holds basic information about a point on the flight track, it has latitude, longitude and altitude in WGS 1984 datum '''

    lat: float  # Degrees of latitude north of the equator, with reference to the WGS84 ellipsoid. For more information see: https://github.com/uastech/standards/blob/master/remoteid/canonical.yaml#L1160
    lng: float  # Degrees of longitude east of the Prime Meridian, with reference to the WGS84 ellipsoid. For more information see: https://github.com/uastech/standards/blob/master/remoteid/canonical.yaml#L1170
    alt: float  # meters in WGS 84, normally calculated as height of ground level in WGS84 and altitude above ground level
    speed: float # speed in m / s 
    bearing: float # forward azimuth for the this and the next point on the track


class AircraftPosition(NamedTuple):
    ''' A object to hold AircraftPosition details for Remote ID purposes, it mataches the RIDAircraftPosition  per the RID standard, for more information see https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1091'''

    lat: float
    lng: float
    alt: float
    accuracy_h: str
    accuracy_v: str
    extrapolated: bool


class AircraftHeight(NamedTuple):
    ''' A object to hold relative altitude for the purposes of Remote ID. For more information see: https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1142 '''

    distance: float
    reference: str


class AircraftState(NamedTuple):
    ''' A object to hold Aircraft state details for remote ID purposes. For more information see the published standard API specification at https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1604 '''

    timestamp: datetime
    operational_status: str
    position: AircraftPosition  # See the definition above
    height: AircraftHeight  # See the definition above
    track: float
    speed: float
    speed_accuracy: str
    vertical_speed: float


class RIDFlight(NamedTuple):
    ''' A object to store details of a remoteID flight '''
    id: str  # ID of the flight for Remote ID purposes, e.g. uss1.JA6kHYCcByQ-6AfU, we for this simulation we use just numeric : https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L943
    aircraft_type: str  # Generic type of aircraft https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1711
    states : List[AircraftState]  # See above for definition


class GridCellFlight(NamedTuple):
    ''' A object to hold details of a grid location and the track within it '''

    bounds: shapely.geometry.polygon.Polygon
    track: List[FlightPoint]


class AdjacentCircularFlightsSimulator():

    ''' A class to generate Flight Paths given a bounding box, this is the main module to generate flight path datasets, the data is generated as latitude / longitude pairs with assoiated with the flights. Additional flight metadata e.g. flight id, altitude, registration number can also be generated '''

    def __init__(self, minx: float, miny: float, maxx: float, maxy: float, utm_zone:str) -> None:
        """ Create an AdjacentCircularFlightsSimulator with the specified bounding box.

            Once these extents are specified, a grid will be created with two rows.  The idea is that multiple flights tracks will be created within the extents.
            Args:
            minx: Western edge of bounding box (degrees longitude)
            maxx: Eastern edge of bounding box (degrees longitude)
            miny: Southern edge of bounding box (degrees latitude)
            maxy: Northern edge of bounding box (degrees latitude)
            utm_zone: UTM Zone string for the location, see https://en.wikipedia.org/wiki/Universal_Transverse_Mercator_coordinate_system to identify the zone for the location. 
            
            Raises:
            ValueError: If bounding box has more area than a 500m x 500m square.
        """

        self.minx = minx
        self.miny = miny
        self.maxx = maxx
        self.maxy = maxy
        self.utm_zone = utm_zone

        self.altitude_agl = 50.0

        self.grid_cells_flight_tracks: List[GridCellFlight] = []

        # This object holds the name and the polygon object of the query boxes. The number of bboxes are controlled by the `box_diagonals` variable
        self.query_bboxes: List[QueryBoundingBox] = []

        self.flight_telemetry: List[List[AircraftState]] = []
        self.bbox_center: List[shapely.geometry.Point] = []

        self.geod = Geod(ellps="WGS84")
        
        self.input_extents_valid()

    def input_extents_valid(self) -> None:
        ''' This method checks if the input extents are valid i.e. small enough, if the extent is too large, we reject them, at the moment it checks for extents less than 500m x 500m square but can be changed as necessary.'''

        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        
        area = abs(self.geod.geometry_area_perimeter(box)[0])

        # Have a area less than 500m x 500m square and more than 300m x 300m square to ensure a 70 m diameter tracks
        if (area) < 250000 and (area) > 90000:
            return
        else:
            raise ValueError("The extents provided are not of the correct size, please provide extents that are less than 500m x 500m and more than 300m x 300m square")

    def generate_query_bboxes(self):
        ''' For the differnet Remote ID checks: No, we need to generate three bounding boxes for the display provider, this method generates the 1 km diagonal length bounding box '''
        # Get center of of the bounding box that is inputted into the generator
        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        center = box.centroid
        self.bbox_center.append(center)
        # Transform to geographic co-ordinates to get areas
        transformer = Transformer.from_crs("epsg:4326", "epsg:3857")
        transformed_x, transformed_y = transformer.transform(center.x, center.y)
        pt = Point(transformed_x, transformed_y)
        # Now we have a point, we can buffer the point and create bounding boxes of the buffer to get the appropriate polygons, more than three boxes can be created, for the tests three will suffice.
        now = datetime.now()

        box_diagonals = [
            {'length': 150, 'name': 'zoomed_in_detail', 'timestamp_after': now + timedelta(seconds=60), 'timestamp_before': now + timedelta(seconds=90)},
            {'length': 380, 'name': "whole_flight_area", 'timestamp_after': now + timedelta(seconds=30), 'timestamp_before': now + timedelta(seconds=60)},
            {'length': 3000, 'name': 'too_large_query', 'timestamp_after': now + timedelta(seconds=10), 'timestamp_before': now + timedelta(seconds=30)}
        ]

        for box_id, box_diagonal in enumerate(box_diagonals):
            # Buffer the point with the appropriate length
            buffer = pt.buffer(box_diagonal['length'])
            buffer_bounds = buffer.bounds
            buffer_bounds_polygon = shapely.geometry.box(buffer_bounds[0], buffer_bounds[1], buffer_bounds[2], buffer_bounds[3])
            buffer_points = zip(buffer_bounds_polygon.exterior.coords.xy[0], buffer_bounds_polygon.exterior.coords.xy[1])
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
            self.query_bboxes.append(QueryBoundingBox(name=box_diagonals[box_id]['name'], shape=buffered_box,
                                                      timestamp_after=box_diagonals[box_id]['timestamp_after'], timestamp_before=box_diagonals[box_id]['timestamp_before']))

    def generate_flight_speed_bearing(self, adjacent_points: List, delta_time_secs: int) -> List[float]:
        ''' A method to generate flight speed, assume that the flight has to traverse two adjecent points in x number of seconds provided, calculating speed in meters / second. It also generates bearing between this and next point, this is used to populate the 'track' paramater in the Aircraft State JSON. '''

        first_point = adjacent_points[0]
        second_point = adjacent_points[1]
        
        fwd_azimuth, back_azimuth, adjacent_point_distance_mts = self.geod.inv(first_point.x, first_point.y, second_point.x, second_point.y)
        
        speed_mts_per_sec = (adjacent_point_distance_mts / delta_time_secs)
        speed_mts_per_sec = float("{:.2f}".format(speed_mts_per_sec))
        
        if fwd_azimuth < 0:
            fwd_azimuth = 360 + fwd_azimuth
            
        return [speed_mts_per_sec, fwd_azimuth]

    def utm_converter(self, shapely_shape: shapely.geometry, inverse:bool=False) -> shapely.geometry.shape:
        ''' A helper function to convert from lat / lon to UTM coordinates for buffering. tracks. This is the UTM projection (https://en.wikipedia.org/wiki/Universal_Transverse_Mercator_coordinate_system), we use Zone 33T which encompasses Switzerland, this zone has to be set for each locale / city. Adapted from https://gis.stackexchange.com/questions/325926/buffering-geometry-with-points-in-wgs84-using-shapely '''
        
        proj = Proj(proj="utm", zone=self.utm_zone, ellps="WGS84", datum="WGS84")
        
        geo_interface = shapely_shape.__geo_interface__
        point_or_polygon = geo_interface['type']
        coordinates = geo_interface['coordinates']
        if point_or_polygon == 'Polygon':
            new_coordinates = [[proj(*point, inverse=inverse) for point in linring] for linring in coordinates]
        elif point_or_polygon == 'Point':
            new_coordinates = proj(*coordinates, inverse=inverse)
        return shapely.geometry.shape({'type': point_or_polygon, 'coordinates': tuple(new_coordinates)})

    def make_json_compatible(self, struct: Any) -> Any:
         if isinstance(struct, tuple) and hasattr(struct, '_asdict'):
             return {k: self.make_json_compatible(v) for k, v in struct._asdict().items()}
         elif isinstance(struct, dict):
             return {k: self.make_json_compatible(v) for k, v in struct.items()}
         elif isinstance(struct, str):
             return struct
         try:
             return [self.make_json_compatible(v) for v in struct]
         except TypeError:
             return struct
    
    def generate_flight_grid_and_path_points(self, altitude_of_ground_level_wgs_84 : float):
        ''' Generate a series of boxes (grid) within the given bounding box to have areas for different flight tracks within each box '''
        # Compute the box where the flights will be created. For a the sample bounds given, over Bern, Switzerland, a division by 2 produces a cell_size of 0.0025212764739985793, a division of 3 is 0.0016808509826657196 and division by 4 0.0012606382369992897. As the cell size goes smaller more number of flights can be accomodated within the grid. For the study area bounds we build a 3x2 box for six flights by creating 3 column 2 row grid.
        N_COLS = 3
        N_ROWS = 2
        cell_size_x = (self.maxx - self.minx)/(N_COLS)  # create three columns
        cell_size_y = (self.maxy - self.miny)/(N_ROWS)  # create two rows
        grid_cells = []
        for u0 in range(0, N_COLS):  # 3 columns
            x0 = self.minx + (u0 * cell_size_x)
            for v0 in range(0, N_ROWS):  # 2 rows
                y0 = self.miny + (v0 * cell_size_y)
                x1 = x0 + cell_size_x
                y1 = y0 + cell_size_y
                grid_cells.append(shapely.geometry.box(x0, y0, x1, y1))

        all_grid_cell_tracks = []
        
        ''' For each of the boxes (grid) allocated to the operator, get the centroid and buffer to generate a flight path. A 70 m radius is provided to have flight paths within each of the boxes '''
        # Iterate over the flight_grid
        for grid_cell in grid_cells:
            center = grid_cell.centroid
            center_utm = self.utm_converter(center)
            buffer_shape_utm = center_utm.buffer(70)
            buffered_path = self.utm_converter(buffer_shape_utm, inverse=True)
            
            altitude = altitude_of_ground_level_wgs_84 + self.altitude_agl  # meters WGS 84
            flight_points_with_altitude = []
            x, y = buffered_path.exterior.coords.xy

            for coord in range(0, len(x)):
                cur_coord = coord
                next_coord = coord + 1 
                next_coord = 0 if next_coord == len(x) else next_coord
                adjacent_points = [Point(x[cur_coord], y[cur_coord]), Point(x[next_coord], y[next_coord])]
                flight_speed, bearing  = self.generate_flight_speed_bearing(adjacent_points=adjacent_points, delta_time_secs= 1)
                
                flight_points_with_altitude.append(FlightPoint(lat = y[coord], lng = x[coord], alt = altitude, speed = flight_speed, bearing = bearing))

            all_grid_cell_tracks.append(GridCellFlight(bounds = grid_cell, track = flight_points_with_altitude))

        self.grid_cells_flight_tracks = all_grid_cell_tracks

    def generate_flight_operator_details(self): 
        ''' This class generates details of flights and operator details for a flight, this data is required for identifying flight, operator and operation  
            Args:
            locale of the format ISO 639 Language Code and - or _ followed by ISO 3166 Country code, not all locales are supported by the Faker module, '''
        
        my_flight_details_generator = details_generator.OperatorFlightDataGenerator()
        
        flight_details = { "serial_number":my_flight_details_generator.generate_serial_number(),"registration_number":my_flight_details_generator.generate_registration_number(),"operation_description":my_flight_details_generator.generate_operation_description()}
        
        operator_details =  { "name":my_flight_details_generator.generate_company_name(), "location":my_flight_details_generator.generate_operator_location(centroid= self.bbox_center[0])}

        flight_operator_details = {'flight_details':flight_details, 'operator_details':operator_details}
        return flight_operator_details

    def generate_rid_state(self, duration=180):
        '''

        This method generates rid_state objects that can be submitted as flight telemetry


        '''
        all_flight_telemetry = {}
        flight_track_details = {}  # Develop a index of flight length and their index
        # Store where on the track the current index is, since the tracks are circular, once the end of the track is reached, the index is reset to 0 to indicate beginning of the track again.
        flight_current_index = {}
        # Get the number of flights
        num_flights = len(self.grid_cells_flight_tracks)
        time_increment_seconds = 1 # the number of seconds it takes to go from one point to next on the track
        now = arrow.now()
        now_isoformat = now.isoformat()
        for i in range(num_flights):
            flight_positions_len = len(self.grid_cells_flight_tracks[i].track)
            
            # in a circular flight pattern increment direction
            angle_increment = (360 / flight_positions_len)
            
            # the resolution of track is 1 degree minimum
            angle_increment = 1.0 if angle_increment == 0.0 else angle_increment

            if i not in flight_track_details:
                flight_track_details[i] = {}
            flight_track_details[i]['track_length'] = flight_positions_len
            
            flight_current_index[i] = 0
            all_flight_telemetry[i]= {}
            all_flight_telemetry[i]['states'] = []

        for j in range(duration):
            if j == 0:
                timestamp = now.shift(seconds=time_increment_seconds)
            else:
                timestamp = timestamp.shift(seconds=time_increment_seconds)
            
                
            timestamp_isoformat = timestamp.isoformat()
            
            for k in range(num_flights): 
                list_end = flight_track_details[k]['track_length'] - \
                    flight_current_index[k]

                if list_end != 1:
                    flight_point = self.grid_cells_flight_tracks[k].track[flight_current_index[k]]
                    aircraft_position = AircraftPosition(lat=flight_point.lat,
                                                         lng=flight_point.lng,
                                                         alt=flight_point.alt,
                                                         accuracy_h="HAUnkown",
                                                         accuracy_v="VAUnknown",
                                                         extrapolated=False,
                                                         )
                    aircraft_height = AircraftHeight(distance=self.altitude_agl, reference="TakeoffLocation")

                    rid_aircraft_state = AircraftState(
                        timestamp=timestamp_isoformat,
                        operational_status="Airborne",
                        position=aircraft_position,
                        height=aircraft_height,
                        track=flight_point.bearing,
                        speed=flight_point.speed,
                        speed_accuracy="SA3mps",
                        vertical_speed=0.0)

                    all_flight_telemetry[k]['states'].append(rid_aircraft_state)
                    
                    flight_current_index[k] += 1
                else:
                    flight_current_index[k] = 0

        telemetery_data_list = []
        for m in range(num_flights):
            
            rid_aircraft_flight = RIDFlight(id=m, aircraft_type="Helicopter", states=all_flight_telemetry[m]['states'])

            rid_aircraft_flight_deserialized = self.make_json_compatible(rid_aircraft_flight)
            telemetery_data_list.append(rid_aircraft_flight_deserialized)
                    
   
        flight_operator_details = self.generate_flight_operator_details()
        
        self.flight_telemetry = {"telemetery_data_list": telemetery_data_list, "reference_time": now_isoformat,'operator_details':flight_operator_details['operator_details'], 'flight_details':flight_operator_details['flight_details']}



class TrackWriter():

    """
        Write the tracks created by AdjacentCircularFlightsSimulator into disk (in the outputs directory) as GeoJSON FeatureCollection 
        Args:
        flight_path_points: A set of flight path points generated by generate_flight_paths_points method in the AdjacentCircularFlightsSimulator class
        bboxes: A set of bounding boxes generated by generate_query_bboxes method in the AdjacentCircularFlightsSimulator class
        country_code: An ISO 3166-1 alpha-3 code for a country

        Outputs: 
        GeoJSON files for bboxes created in the `test_definitions/{country_code}` folder 


    """

    def __init__(self, grid_tracks: List[GridCellFlight], bboxes: List[QueryBoundingBox], country_code='CHE') -> None:
        ''' This class uses the same output directory as the AdjacentCircularFlightsSimulator class and requires the path points (Tracks) and the bounding boxes from that class.

        '''

        self.grid_cells_flight_tracks = grid_tracks
        self.bboxes = bboxes
        self.country_code = country_code
        self.output_directory = Path('test_definitions', self.country_code)
        # Create test_definition directory if it does not exist
        self.output_directory.mkdir(parents=True, exist_ok=True)
        self.output_subdirectories = (Path(self.output_directory, 'tracks'), Path(self.output_directory, 'query_bboxes'))

        for output_subdirectory in self.output_subdirectories:
            output_subdirectory.mkdir(parents=True, exist_ok=True)

    def write_bboxes(self):
        ''' This module writes the bboxes as a GeoJSON FeatureCollection '''
        for buffered_bbox_details in self.bboxes:

            features = json.dumps({'type': 'Feature', 'properties': {"timestamp_before": buffered_bbox_details.timestamp_before.isoformat(), "timestamp_after": buffered_bbox_details.timestamp_after.isoformat()}, 'geometry': shapely.geometry.mapping(buffered_bbox_details.shape)})
            bbox_file_name = 'box_%s.geojson' % buffered_bbox_details.name
            
            bbox_output_path = self.output_subdirectories[1] / bbox_file_name

            with open(bbox_output_path, 'w') as f:
                f.write(features)

    def write_tracks(self):
        ''' This module writes tracks as a GeoJSON FeatureCollection (of Point Feature) for use in other software '''

        flight_point_current_index = {}
        num_flights = len(self.grid_cells_flight_tracks)
        for i in range(num_flights):
            flight_point_current_index[i] = 0

        for track_id, grid_cell_flight_track in enumerate(self.grid_cells_flight_tracks):
            feature_collection = {"type": "FeatureCollection", "features": []}
            point_collection = []
            for cur_track_point in grid_cell_flight_track.track:
                p = Point((cur_track_point.lng, cur_track_point.lat, cur_track_point.alt))
                point_collection.append(p)

            line = LineString(point_collection)
            line_feature = {'type': 'Feature', 'properties': {}, 'geometry': shapely.geometry.mapping(line)}                
            feature_collection['features'].append(line_feature)
            path_file_name = 'track_%s.geojson' % str(track_id + 1)  # Avoid Zero based numbering
            
            tracks_file_path = self.output_subdirectories[0] / path_file_name
            with open(tracks_file_path, 'w') as f:
                f.write(json.dumps(feature_collection))


class RIDAircraftStateWriter():

    """Write tracks in RIDAircraftState object to disk (refer. https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1604)

    """

    def __init__(self, flight_telemetry, country_code='che') -> None:
        """ Atleast single flight points array is necessary and a ouptut directory  
        Args:
        flight_telemetry: 
        country_code: An ISO 3166-1 alpha-3 code for a country, this is used to create a sub-directory to store output.

        Outputs: 
        A JSON datastructure as a file that can be submitted as a part of the test harness to a USS that implements the automatic remote id testing interface. 

        """

        self.flight_telemetry = flight_telemetry
        self.country_code = country_code
        self.flight_telemetry_check()

        self.output_directory = Path('test_definitions', self.country_code)
        # Create test_definition directory if it does not exist
        self.output_directory.mkdir(parents=True, exist_ok=True)
        self.output_subdirectories = (Path(self.output_directory, 'aircraft_states'),)
        for output_subdirectory in self.output_subdirectories:
            output_subdirectory.mkdir(parents=True, exist_ok=True)

    def flight_telemetry_check(self) -> None:
        ''' Check if atleast one track is provided, if no tracks are provided, then RIDAircraftState and Test JSON cannot be generated.'''

        # Empty flight points cannot be converted to a Aircraft State, check if the list has
        if (self.flight_telemetry['telemetery_data_list']):
            return
        else:
            raise ValueError("At least one flight track is necessary to create a AircraftState and a test JSON, please generate the tracks first using AdjacentCircularFlightsSimulator class")

    def write_rid_state(self):
        ''' This method iterates over flight tracks and geneates AircraftState JSON objects and writes to disk in the test_definitions folder, these files can be used to submit the data in the test harness '''

        reference_time = self.flight_telemetry['reference_time']
        flight_details = self.flight_telemetry['flight_details']
        operator_details = self.flight_telemetry['operator_details']

        for flight_id, single_flight_telemetry_data in enumerate(self.flight_telemetry['telemetery_data_list']):

            rid_test_file_name = 'flight_' + str(flight_id + 1) + '_rid_aircraft_state' + '.json' # Add 1 to avoid zero based numbering            
            
            rid_test_file_path = self.output_subdirectories[0] / rid_test_file_name
            flight_telemetry_data = {'reference_time':reference_time, 'flight_telemetry':single_flight_telemetry_data,'flight_details':flight_details, 'operator_details':operator_details }
            with open(rid_test_file_path, 'w') as f:
                f.write(json.dumps(flight_telemetry_data))


if __name__ == '__main__':
    # TODO: accept these parameters as values so that other locations can be supplied
    my_path_generator = AdjacentCircularFlightsSimulator(minx=7.4735784530639648, miny=46.9746744128218410, maxx=7.4786210060119620, maxy=46.9776318195799121, utm_zone='32T')    
    altitude_of_ground_level_wgs_84 = 570 # height of the geoid above the WGS84 ellipsoid (using EGM 96) for Bern, rom https://geographiclib.sourceforge.io/cgi-bin/GeoidEval?input=46%B056%26%238242%3B53%26%238243%3BN+7%B026%26%238242%3B51%26%238243%3BE&option=Submit
    COUNTRY_CODE = 'che'
    flight_points = []
    query_bboxes = []

    my_path_generator.generate_flight_grid_and_path_points(altitude_of_ground_level_wgs_84 = altitude_of_ground_level_wgs_84)
    my_path_generator.generate_query_bboxes()

    grid_tracks = my_path_generator.grid_cells_flight_tracks

    my_path_generator.generate_rid_state(duration=180)
    flight_telemetry = my_path_generator.flight_telemetry

    query_bboxes = my_path_generator.query_bboxes

    my_track_writer = TrackWriter(grid_tracks=grid_tracks, bboxes=query_bboxes, country_code=COUNTRY_CODE)
    my_track_writer.write_bboxes()
    my_track_writer.write_tracks()

    my_state_generator = RIDAircraftStateWriter(flight_telemetry=flight_telemetry, country_code=COUNTRY_CODE)
    my_state_generator.write_rid_state()
