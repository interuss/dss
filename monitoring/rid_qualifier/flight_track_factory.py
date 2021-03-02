from shapely.geometry import Point, Polygon
import shapely.geometry
from pyproj import Geod, Transformer
import json
from pathlib import Path
from typing import List, ClassVar

class AdjacentCircularFlightPathsGenerator():

    ''' A class to generate Flight Paths given a bounding box, this is the main module to generate flight path datasets, the data is generated as latitude / longitude pairs with assoiated with the flights. Additional flight metadata e.g. flight id, altitude, registration number can also be generated '''


    flight_points: ClassVar[List] = []   # This is a object that containts multiple lists of flight tracks as points, in the latitude, longitude, altitude in tuple format. Depending on how the grid is generated in this case 3 columns and 2 rows with six flight tracks there will be six lists in this object
    flight_grid: ClassVar[List] = [] # This object holds the polygon objects for the different grid cells within the bounding box. 
    query_bboxes: ClassVar[List] = [] # This object holds the name and the polygon object of the query boxes. The number of bboxes are controlled by the `box_diagonals` variable

    def __init__(self, minx: float, miny: float, maxx: float, maxy: float) -> None: 
        # A bounding box (extents) is necessary, here the minx and miny detail the SE corner where minx is the longitude and miny is the latitude and maxx and maxy is the NW corner and maxx is the longitude and maxy is the latitude. Once these extents are specified, a grid needs to be created with two rows. The max_cols parameter specifies the desired columns inside the extents. The idea is that multiple flights will be created within the extents
        self.minx = minx 
        self.miny = miny
        self.maxx = maxx
        self.maxy = maxy                
        
        if not self.input_extents_valid():
            raise ValueError

    def input_extents_valid(self) :

        ''' This method checks if the input extents are valid i.e. small enough, if the extent is too large, we reject them, at the moment it checks for extents less than 500m x 500m square but can be changed as necessary.''' 

        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        geod = Geod(ellps="WGS84")
        area = abs(geod.geometry_area_perimeter(box)[0])
        if (area) < 250000: # Have a area less than 500m x 500m square
            return True
        else: 
            return False
        
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
        box_diagonals = {1:{'length':150, 'name':'zoomed_in_detail'},2:{'length':380, 'name':"whole_flight_area"},3:{'length':3000, 'name':'too_large_query'}}

        for box_id, box_diagonal in box_diagonals.items():
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
            self.query_bboxes.append({'name':box_diagonals[box_id]['name'], 'shape': buffered_box})
            

    def generate_flight_grid(self):
        ''' Generate a series of boxes within the bounding box to have areas for different flights '''
        # Compute the box where the flights will be created. For a the sample bounds given, over Bern, Switzerland, a division by 2 produces a cell_size of 0.0025212764739985793, a division of 3 is 0.0016808509826657196 and division by 4 0.0012606382369992897. As the cell size goes smaller more number of flights can be accomodated within the grid. For the study area bounds we build a 3x2 box for six flights by creating 3 column 2 row grid. 

        cell_size_x = (self.maxx - self.minx)/(3) # create three columns
        cell_size_y = (self.maxy - self.miny)/(2) # create two rows
        grid_cells = [] 
        for u0 in range(0, 3):  # 3 columns           
            x0 =  self.minx + (u0 * cell_size_x)            
            for v0 in range(0,2): # 2 rows
                y0 = self.miny + (v0 *cell_size_y)
                x1 = x0 + cell_size_x
                y1 = y0 + cell_size_y
                grid_cells.append(shapely.geometry.box(x0, y0, x1, y1))                
        
        self.flight_grid = grid_cells
        

    def generate_flight_paths_points(self):
        ''' For each of the boxes allocated to the operator, get the centroid and buffer to get a flight path. A 75 m radius is provided to have flight paths within each of the boxes '''
        # Iterate over the flight_grid
        for grid_cell in self.flight_grid:
            center = grid_cell.centroid
            ## Transfrom to buffer 100 m diameter circle on which the drone will fly            
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
                flight_points_with_altitude.append((x[coord], y[coord], altitude))

            # Build a list of points so that they can be fed to the sim and outputted. 
            self.flight_points.append(flight_points_with_altitude)


class TrackWriter():
    ''' A class to write data geometry as files in geographical data format''' 

    def __init__(self, path_points, bboxes) -> None:

        self.flight_path_points = path_points
        self.bboxes = bboxes
        self.loop_counter = 0
        
        output_directory = Path('outputs').mkdir(parents=True, exist_ok=True) # Create outputs directory if it does not exist

    def write_bboxes(self):
        for buffered_bbox_details in self.bboxes:
            
            features = json.dumps({'type': 'Feature', 'properties': {}, 'geometry': shapely.geometry.mapping(buffered_bbox_details['shape'])})
            with open('outputs/box_%s.geojson'% buffered_bbox_details['name'],'w') as f:
                f.write(features)
                
    def write_tracks(self):
        ''' This module writes tracks as a GeoJSON / KML for use in other software '''       
        
        flight_point_lenghts = {}
        flight_point_current_index = {}
        num_flights = len(self.flight_path_points)
        for i in range(num_flights):
            flight_point_lenghts[i]= len(self.flight_path_points[i])
            flight_point_current_index[i] = 0
            
        for track_id, flight_track in enumerate(self.flight_path_points):
            feature_collection = {"type":"FeatureCollection", "features": []}
            for cur_track_point in flight_track:                
                p = Point(cur_track_point)
                point_feature = {'type': 'Feature', 'properties': {}, 'geometry': shapely.geometry.mapping(p)}                
                feature_collection['features'].append(point_feature)

            with open('outputs/track_%s.geojson'% str(track_id+1),'w') as f:
                f.write(json.dumps(feature_collection))

            

if __name__ == '__main__':
    #TODO: accept these parameters as values so that other locations can be supplied
    my_path_generator = AdjacentCircularFlightPathsGenerator(minx = 7.4735784530639648, miny = 46.9746744128218410, maxx = 7.4786210060119620, maxy= 46.9776318195799121)
    flight_points = []
    query_bboxes = []
    try:
        assert my_path_generator.input_extents_valid()
    except AssertionError as ae:
        print("Extents are too large, please have extents less than 500m x 500m square")
    else:
        my_path_generator.generate_flight_grid()
        my_path_generator.generate_flight_paths_points()
        flight_points = my_path_generator.flight_points        
        my_path_generator.generate_query_bboxes()
        query_bboxes= my_path_generator.query_bboxes
    # Start the flight simulator    
    my_track_writer = TrackWriter(path_points = flight_points,bboxes=query_bboxes)
    my_track_writer.write_bboxes()
    my_track_writer.write_tracks()

