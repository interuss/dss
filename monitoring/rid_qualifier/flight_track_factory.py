from binascii import b2a_base64
import numpy as np
from shapely.geometry import Point, Polygon
import shapely.geometry
from pyproj import Geod, Transformer
import json
import time

class FlightPathGenerator():

    ''' A class to generate Flight Paths given a bounding box, this is the main module to generate flight path datasets, the data is generated as latitude / longitude pairs with assoiated with the flights. Additional flight metadata e.g. flight id, altitude, registration number can also be generated '''

    crs = "+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs"    
    flight_paths = []
    flight_points = []
    flight_grid  = []
    query_bboxes = []

    def __init__(self, minx, miny, maxx, maxy, max_cols = 4): 
        # By default reference geojsons will not be created but by using the crate_geojson flag you can output Geojsons, in the geojson directory.
        self.minx = minx
        self.miny = miny
        self.maxx = maxx
        self.maxy = maxy                
        self.max_cols = max_cols

    def input_extents_valid(self):

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
        box_diagonals = {1:440,2:900,3:3000}

        for box_id, box_diagonal in box_diagonals.items():
            # Buffer the point with the appropriate length
            buffer = pt.buffer(box_diagonal)                   
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
            buffered_box_bounds = buffered_box.bounds
            self.query_bboxes.append(buffered_box_bounds)
            

    def generate_flight_grid(self):
        ''' Generate a series of boxes within the bounding box to have areas for different flights '''

        cell_size = (self.maxx-self.minx)/(self.max_cols//2)        
        grid_cells = []
        for x0 in np.arange(self.minx, self.maxx+cell_size, cell_size ):
            for y0 in np.arange(self.miny, self.miny+cell_size, cell_size):
                x1 = x0-cell_size
                y1 = y0+cell_size
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
            buffer = pt.buffer(75)
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

            self.flight_paths.append(buffered_path)
            
            # Build a list of points so that they can be fed to the sim and outputted. 
            self.flight_points.append(list(zip(*buffered_path.exterior.coords.xy)))


class TrackWriter():
    ''' A basic simulator to loop through flight paths and output to console, or CSV or HTTP requests positions of flights ''' 

    def __init__(self, path_points, bboxes):
        self.flight_path_points = path_points
        self.bboxes = bboxes
        self.loop_counter = 0

    def write_bboxes(self):
        for box_id, buffered_bbox in enumerate(self.bboxes): 
            features = json.dumps({'type': 'Feature', 'properties': {}, 'geometry': shapely.geometry.mapping(buffered_bbox)})
            with open('outputs/box_%s.geojson'% box_id,'w') as f:
                f.write(features)
                
    def write_tracks(self):
        ''' This module writes tracks as a GeoJSON / KML for use in other software ''' 
        pass
                    
    def write_track_payload(self, duration = 180):
        ''' This starts the simulation for 3 minutes and prints flight positions that can be send to the harness every 3 seconds '''        
        # Get the length of flight each paths, this is useful to loop back the index at the end of lists, so that at the end of the path, it goes back to the beginning of the list while timestep is counted down.
        flight_point_lenghts = {}
        flight_point_current_index = {}
        num_flights = len(self.flight_path_points)
        for i in range(num_flights):
            flight_point_lenghts[i]= len(self.flight_path_points[i])
            flight_point_current_index[i] = 0
        
        for j in range(duration):
            for k in range(num_flights):
                list_end = flight_point_lenghts[k] - flight_point_current_index[k]            
                if list_end != 1:             
                    print("Flight %s, timestep %s"% (k, j))   
                    print(self.flight_path_points[k][flight_point_current_index[k]])
                    flight_point_current_index[k]+= 1
                else:
                    flight_point_current_index[k] = 0
            time.sleep(2)
                    

if __name__ == '__main__':
    
    my_path_generator = FlightPathGenerator(minx = 7.4735784530639648, miny = 46.9746744128218410, maxx = 7.4786210060119620, maxy= 46.9776318195799121)
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
    my_track_writer = TrackWriter(flight_points)
    my_track_writer.write_bboxes(query_bboxes)
    my_track_writer.write_track_payload(query_bboxes)
    my_track_writer.write_tracks(query_bboxes)
    

        