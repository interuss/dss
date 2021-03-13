from shapely.geometry import Point, Polygon
import shapely.geometry
from pyproj import Geod, Transformer
import json
from pathlib import Path
from typing import List, NamedTuple
import arrow
import datetime
from datetime import datetime, timedelta
import uuid

class QueryBoundingBox(NamedTuple):
    ''' This is the object that stores details of query bounding box '''
    name: str
    shape: Polygon
    timestamp_before: timedelta
    timestamp_after: timedelta
    
class FlightPoint(NamedTuple):
    ''' This is the object that stores details of query bounding box '''
    lat: float
    lng: float
    alt: float

class AircraftPosition(NamedTuple):
    ''' A object to hold AircraftPosition for Remote ID purposes. For more information see, the definition in the standard: https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1091'''

    lat : float
    lng : float
    alt : float
    accuracy_h : str
    accuracy_v : str
    extrapolated: int
    pressure_altitude : int

class AdjacentCircularFlightPathsGenerator():

    ''' A class to generate Flight Paths given a bounding box, this is the main module to generate flight path datasets, the data is generated as latitude / longitude pairs with assoiated with the flights. Additional flight metadata e.g. flight id, altitude, registration number can also be generated '''

    def __init__(self, minx: float, miny: float, maxx: float, maxy: float) -> None: 

        """ Create an AdjacentCircularFlightPathsGenerator with the specified bounding box.

            Once these extents are specified, a grid will be created with two rows.  The idea is that multiple flights tracks will be created within the extents.
            Args:
            minx: Western edge of bounding box (degrees longitude)
            maxx: Eastern edge of bounding box (degrees longitude)
            miny: Southern edge of bounding box (degrees latitude)
            maxy: Northern edge of bounding box (degrees latitude)

            Raises:
            ValueError: If bounding box has more area than a 500m x 500m square.
        """

        self.minx = minx 
        self.miny = miny
        self.maxx = maxx
        self.maxy = maxy                
        
        self.flight_points: List[FlightPoint] = []   # This is a object that containts multiple lists of flight tracks as points, in the latitude, longitude, altitude in tuple format. Depending on how the grid is generated in this case 3 columns and 2 rows with six flight tracks there will be six lists in this object
        self.flight_grid: List[shapely.geometry.polygon.Polygon] = [] # This object holds the polygon objects for the different grid cells within the bounding box. 
        self.query_bboxes: List[QueryBoundingBox] = [] # This object holds the name and the polygon object of the query boxes. The number of bboxes are controlled by the `box_diagonals` variable

        self.input_extents_valid()
        

    def input_extents_valid(self) -> None:

        ''' This method checks if the input extents are valid i.e. small enough, if the extent is too large, we reject them, at the moment it checks for extents less than 500m x 500m square but can be changed as necessary.''' 

        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        geod = Geod(ellps="WGS84")
        area = abs(geod.geometry_area_perimeter(box)[0])
        if (area) < 250000: # Have a area less than 500m x 500m square
            return
        else: 
            raise ValueError("The extents provided are too large, please provide extents that are less than 500m x 500m square")
        
    def generate_query_bboxes(self):

        ''' For the differnet Remote ID checks: No, we need to generate three bounding boxes for the display provider, this method generates the 1 km diagonal length bounding box '''
        # Get center of of the bounding box that is inputted into the generator
        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        center = box.centroid
        # Transform to geographic co-ordinates to get areas
        transformer = Transformer.from_crs("epsg:4326", "epsg:3857")
        transformed_x,transformed_y = transformer.transform(center.x, center.y)
        pt = Point(transformed_x,transformed_y)
        # Now we have a point, we can buffer the point and create bounding boxes of the buffer to get the appropriate polygons, more than three boxes can be created, for the tests three will suffice.
        now = datetime.now()
        
        box_diagonals = [
            {'length':150, 'name':'zoomed_in_detail', 'timestamp_after':now + timedelta(seconds=60),'timestamp_before':now + timedelta(seconds=90)},
            {'length':380, 'name':"whole_flight_area",'timestamp_after':now + timedelta(seconds=30),'timestamp_before':now + timedelta(seconds=60)},
            {'length':3000, 'name':'too_large_query','timestamp_after':now + timedelta(seconds=10),'timestamp_before':now + timedelta(seconds=30)}
            ]

        for box_id, box_diagonal in enumerate(box_diagonals):
            # Buffer the point with the appropriate length
            buffer = pt.buffer(box_diagonal['length'])                   
            buffer_bounds = buffer.bounds
            buffer_bounds_polygon  =  shapely.geometry.box(buffer_bounds[0], buffer_bounds[1], buffer_bounds[2], buffer_bounds[3])
            buffer_points =  zip(buffer_bounds_polygon.exterior.coords.xy[0], buffer_bounds_polygon.exterior.coords.xy[1])
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
            self.query_bboxes.append(QueryBoundingBox(name=box_diagonals[box_id]['name'], shape=buffered_box, timestamp_after = box_diagonals[box_id]['timestamp_after'], timestamp_before = box_diagonals[box_id]['timestamp_before']))
            

    def generate_flight_grid(self):

        ''' Generate a series of boxes within the bounding box to have areas for different flights '''
        # Compute the box where the flights will be created. For a the sample bounds given, over Bern, Switzerland, a division by 2 produces a cell_size of 0.0025212764739985793, a division of 3 is 0.0016808509826657196 and division by 4 0.0012606382369992897. As the cell size goes smaller more number of flights can be accomodated within the grid. For the study area bounds we build a 3x2 box for six flights by creating 3 column 2 row grid. 
        N_COLS = 3
        N_ROWS = 2
        cell_size_x = (self.maxx - self.minx)/(N_COLS) # create three columns
        cell_size_y = (self.maxy - self.miny)/(N_ROWS) # create two rows
        grid_cells = [] 
        for u0 in range(0, N_COLS):  # 3 columns           
            x0 =  self.minx + (u0 * cell_size_x)            
            for v0 in range(0,N_ROWS): # 2 rows
                y0 = self.miny + (v0 *cell_size_y)
                x1 = x0 + cell_size_x
                y1 = y0 + cell_size_y
                grid_cells.append(shapely.geometry.box(x0, y0, x1, y1))                
        
        self.flight_grid = grid_cells
        

    def generate_flight_paths_points(self):

        ''' For each of the boxes allocated to the operator, get the centroid and buffer to get a flight path. A 70 m radius is provided to have flight paths within each of the boxes '''

        # Iterate over the flight_grid
        for grid_cell in self.flight_grid:
            center = grid_cell.centroid
            ## Transfrom to buffer 140 m diameter circle on which the drone will fly            
            transformer = Transformer.from_crs("epsg:4326", "epsg:3857")
            transformed_x,transformed_y = transformer.transform(center.x, center.y)
            pt = Point(transformed_x,transformed_y)
            buffer = pt.buffer(70) # build a buffer so that the radius is 70m for the track
            buffer_points =  zip(buffer.exterior.coords.xy[0], buffer.exterior.coords.xy[1])
            proj_buffer_points = []
            # reproject back to ESPG 4326
            transformer2 = Transformer.from_crs("epsg:3857", "epsg:4326")
            for point in buffer_points:
                x = point[0]
                y = point[1]
                x, y = transformer2.transform(x, y)
                proj_buffer_points.append((x, y))
            buffered_path = Polygon(proj_buffer_points)
            altitude = 70.0
            flight_points_with_altitude = []
            x, y = buffered_path.exterior.coords.xy
            for coord in range(0,len(x)):
                flight_points_with_altitude.append(FlightPoint(lat = x[coord],lng = y[coord],alt= altitude))

            # Build a list of points so that they can be fed to the sim and outputted. 
            self.flight_points.append(flight_points_with_altitude)


class TrackWriter():

    """
        Write the tracks created by AdjacentCircularFlightPathsGenerator into disk (in the outputs directory) as GeoJSON FeatureCollection 
        Args:
        flight_path_points: A set of flight path points generated by generate_flight_paths_points method in the AdjacentCircularFlightPathsGenerator class
        bboxes: A set of bounding boxes generated by generate_query_bboxes method in the AdjacentCircularFlightPathsGenerator class
        country_code: An ISO 3166-1 alpha-3 code for a country

        Outputs: 
        GeoJSON files for bboxes created in the `test_definitions/{country_code}` folder 
        
        
    """

    def __init__(self, path_points:  List[List[FlightPoint]], bboxes: List[QueryBoundingBox], country_code='CHE') -> None:
        ''' This class uses the same output directory as the AdjacentCircularFlightPathsGenerator class and requires the path points (Tracks) and the bounding boxes from that class.

        '''
        
        self.flight_path_points = path_points
        self.bboxes = bboxes
        self.country_code = country_code        
        self.output_directory = Path('test_definitions', self.country_code)
        self.output_directory.mkdir(parents=True, exist_ok=True) # Create test_definition directory if it does not exist

    def write_bboxes(self):

        ''' This module writes the bboxes as a GeoJSON FeatureCollection '''
        for buffered_bbox_details in self.bboxes:
            
            features = json.dumps({'type': 'Feature', 'properties': {"timestamp_before":buffered_bbox_details.timestamp_before.isoformat(), "timestamp_after":buffered_bbox_details.timestamp_after.isoformat()}, 'geometry': shapely.geometry.mapping(buffered_bbox_details.shape)})
            bbox_file_name = 'box_%s.geojson'% buffered_bbox_details.name
            bbox_output_path = self.output_directory / bbox_file_name

            with open(bbox_output_path,'w') as f:
                f.write(features)
                
    def write_tracks(self):

        ''' This module writes tracks as a GeoJSON FeatureCollection (of Point Feature) for use in other software '''       
        
        flight_point_lenghts = {}
        flight_point_current_index = {}
        num_flights = len(self.flight_path_points)
        for i in range(num_flights):
            flight_point_lenghts[i]= len(self.flight_path_points[i])
            flight_point_current_index[i] = 0
            
        for track_id, flight_track in enumerate(self.flight_path_points):
            feature_collection = {"type":"FeatureCollection", "features": []}
            for cur_track_point in flight_track:                
                p = Point((cur_track_point.lat, cur_track_point.lng, cur_track_point.alt))
                point_feature = {'type': 'Feature', 'properties': {}, 'geometry': shapely.geometry.mapping(p)}                
                feature_collection['features'].append(point_feature)

            path_file_name = 'track_%s.geojson'% str(track_id+1)
            tracks_file_path = self.output_directory / path_file_name
            with open(tracks_file_path,'w') as f:
                f.write(json.dumps(feature_collection))

class RIDAircraftStateWriter():

    """Convert the tracks created by AdjacentCircularFlightPathsGenerator into RIDAircraftState object (refer. https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1604)

       
    """

    def __init__(self, flight_points, country_code='che') -> None:

        """ Atleast single flight points array is necessary and a ouptut directory  
        Args:
        flight_points: A list of flight points each in FlightPoint format, generated from generate_flight_paths_points method in the AdjacentCircularFlightPathsGenerator method.
        country_code: An ISO 3166-1 alpha-3 code for a country, this is used to create a sub-directory to store output.
        
        Outputs: 
        A JSON datastructure as a file that can be submitted as a part of the test harness to a USS that implements the automatic remote id testing interface. 
        
        """

        self.flight_points = flight_points
        self.country_code = country_code
        self.flight_points_check()

        self.output_directory = Path('test_definitions', self.country_code)
        self.output_directory.mkdir(parents=True, exist_ok=True) # Create test_definition directory if it does not exist


    def flight_points_check(self) -> None:

        ''' Check if atleast one track is provided, if no tracks are provided, then RIDAircraftState and Test JSON cannot be generated.''' 

        if (self.flight_points): # Empty flight points cannot be converted to a Aircraft State, check if the list has 
            return
        else:
            raise ValueError("At least one flight track is necessary to create a AircraftState and a test JSON, please generate the tracks first using AdjacentCircularFlightPathsGenerator class")

    def write_rid_state(self, duration = 180):

        ''' This method iterates over flight tracks and geneates AircraftState JSON objects and writes to disk in the test_definitions folder, these files can be used to submit the data in the test harness '''

        flight_lenghts = {} # Develop a index of flight length and their index
        flight_current_index = {} # Store where on the track the current index is, since the tracks are circular, once the end of the track is reached, the index is reset to 0 to indicate beginning of the track again. 
        num_flights = len(self.flight_points) # Get the number of flights 
        now = arrow.now() 
        flight_telemetry = {}
        for i in range(num_flights):
            flight_lenghts[i]= len(self.flight_points[i])
            flight_current_index[i] = 0
            flight_telemetry[i] = []
        
        for j in range(duration):
            if j == 0:
                timestamp = now.shift(seconds = 2)
            else:
                timestamp = timestamp.shift(seconds = 2)
            seconds_diff = (now - timestamp).total_seconds()
            
            for k in range(num_flights):
                list_end = flight_lenghts[k] - flight_current_index[k]            
                if list_end != 1:             
                    flight_point = self.flight_points[k][flight_current_index[k]]
                    aircraft_position = AircraftPosition(lat = flight_point.lat, 
                                                         lng = flight_point.lng, 
                                                         alt = flight_point.alt, 
                                                         accuracy_h= "HAUnkown", 
                                                         accuracy_v = "VAUnknown", 
                                                         extrapolated = 1, 
                                                         pressure_altitude = 0)

                    rid_aircraft_state = {'id':k, 
                                          "aircraft_type":"NotDeclared", 
                                          "current_state":
                                              {"timestamp": seconds_diff,
                                               "operational_status":"Undeclared", 
                                               "position":
                                                   {"lat":aircraft_position.lat, 
                                                    "lng":aircraft_position.lng, 
                                                    "alt":aircraft_position.alt, 
                                                    "accuracy_h": aircraft_position.accuracy_h, "accuracy_v":aircraft_position.accuracy_v, "extrapolated":aircraft_position.extrapolated, 
                                                    "pressure_altitude": aircraft_position.pressure_altitude
                                                    }, 
                                                "height": {"distance":70, "reference": "TakeoffLocation"},
                                                "track":45,
                                                "speed":1.9, 
                                                "speed_accuracy":"SA3mps", 
                                                "vertical_speed":0.2,
                                                "group_radius":20, 
                                                "group_ceiling": 80, 
                                                "group_floor": 10, 
                                                "group_count": 1, 
                                                "group_time_start": seconds_diff, 
                                                "group_time_end": seconds_diff
                                                }
                                            } # In the RID Aircraft state object the only object to update is the "seconds_diff" in the rid_aircraft_state
                    flight_telemetry[k].append(rid_aircraft_state)
                    flight_current_index[k]+= 1
                else:
                    flight_current_index[k] = 0


        for flight_id, single_flight_telemetry in flight_telemetry.items():
            flight_telemetry_data = single_flight_telemetry[flight_id]
            rid_test_file_name = 'flight_' + str(flight_id) + '_rid_aircraft_state' + '.json'
            rid_test_file_path = self.output_directory / rid_test_file_name
            
            with open(rid_test_file_path,'w') as f:
                f.write(json.dumps(flight_telemetry_data))


        
if __name__ == '__main__':
    #TODO: accept these parameters as values so that other locations can be supplied
    my_path_generator = AdjacentCircularFlightPathsGenerator(minx = 7.4735784530639648, miny = 46.9746744128218410, maxx = 7.4786210060119620, maxy= 46.9776318195799121)
    COUNTRY_CODE = 'che'
    flight_points = []
    query_bboxes = []

    my_path_generator.generate_flight_grid()
    my_path_generator.generate_flight_paths_points()    
    my_path_generator.generate_query_bboxes()

    flight_points = my_path_generator.flight_points    
    query_bboxes = my_path_generator.query_bboxes
  
    my_track_writer = TrackWriter(path_points = flight_points,bboxes=query_bboxes, country_code = COUNTRY_CODE)
    my_track_writer.write_bboxes()
    my_track_writer.write_tracks()

    my_state_generator = RIDAircraftStateWriter(flight_points = flight_points, country_code= COUNTRY_CODE)
    my_state_generator.write_rid_state()
