# A file to generate Flight Records from KML.
import json
import s2sphere
import pprint
from shapely.geometry import LineString
from monitoring.rid_qualifier.utils import QueryBoundingBox, FlightPoint, GridCellFlight, FlightDetails, FullFlightRecord
from monitoring.monitorlib.rid import RIDHeight, RIDAircraftState, RIDAircraftPosition, RIDFlightDetails
from monitoring.rid_qualifier import operator_flight_details_generator as details_generator
from monitoring.monitorlib.geo import flatten, unflatten

def get_flight_details():
    # TODO: To be fetched from KML.
    pass

def get_flight_coordinates():
    # TODO: To be fetched from KML.
    # Hardcoded for now.
    return [
        (-77.55585786189934,39.07946259742567,116),
        (-77.55585786189934,39.07946259742567,140),
        (-77.55590933873647,39.07942136290154,140),
        (-77.55640842624625,39.07909259546974,140),
        (-77.55649533040686,39.07907891407359,140),
        (-77.55656511299719,39.07907891839287,140),
        (-77.55665246960693,39.07907899099997,140),
        (-77.55670489443796,39.07907904512823,140),
        (-77.55680975038514,39.07907915754353,140),
        (-77.55689713574773,39.07907925110334,140),
        (-77.55696739403606,39.07909300702381,140),
        (-77.55703851497563,39.07912071630042,140),
        (-77.55707472267461,39.07916199005058,140),
        (-77.55712912144568,39.0792305884167,140),
        (-77.55716620171791,39.07928583292331,140),
        (-77.55720330828555,39.07934111689317,140),
        (-77.55722227978741,39.07939600625936,140),
        (-77.55725765375473,39.0794501116203,140),
        (-77.55727627514482,39.07950482351895,140),
        (-77.55729490448188,39.07955955731004,140),
        (-77.5573137222033,39.07961467471912,140),
        (-77.55733256972054,39.0796695953988,140),
        (-77.55733383683219,39.07972447970517,140),
        (-77.55733483647411,39.07976568977167,140),
        (-77.55733655271884,39.07983444497697,140),
        (-77.55733835028354,39.07990323030949,140),
        (-77.55734066746388,39.07997247920742,140),
        (-77.55734188188015,39.08002742894315,140),
        (-77.5573429671918,39.08006880680381,140),
        (-77.55734477441732,39.08013780308103,140),
        (-77.55734661460282,39.08020683233811,140),
        (-77.55734812343894,39.08026213198261,140),
        (-77.55734925527744,39.08030362198904,140),
        (-77.55735189603955,39.08040049472836,140),
        (-77.55735473821215,39.08049758267686,140),
        (-77.55733867260282,39.0805529976702,140),
        (-77.55732254553646,39.08060840544565,140),
        (-77.55730641159623,39.08066383949771,140),
        (-77.55729063887122,39.08073317882214,140),
        (-77.55727411914556,39.08077478496765,140),
        (-77.55727596510425,39.08084423278386,140),
        (-77.55727744481189,39.08089981792946,140),
        (-77.55726127062646,39.0809553416028,140),
        (-77.55724467575784,39.08099696856388,140),
        (-77.5572108168311,39.08105249752664,140),
        (-77.55721285835257,39.08113590549704,140),
        (-77.55721431069817,39.08119160561851,140),
        (-77.55721605619564,39.08126121254306,140),
        (-77.55721780195692,39.08133085977851,140),
        (-77.55721920016256,39.08138660663652,140),
        (-77.5572209478924,39.08145632573775,140),
        (-77.55723964351446,39.08149819361299,140),
        (-77.55725940226051,39.08158195475191,140),
        (-77.55727882094916,39.08165180407102,140),
        (-77.5573155596023,39.08170773336774,140),
        (-77.5573537218297,39.08181964183621,140),
        (-77.55739126508709,39.08190366249713,140),
        (-77.55741075481436,39.08197370309114,140),
    ]

def get_distance_travelled():
    # TODO: should be calculated based on speed m/s, hardcoding to 2m/s for now.
    return 2

def get_distance_between_two_points(flatten_point1, flatten_point2):
    print(f'flatten_point1: {flatten_point1}, flatten_point2: {flatten_point2}')
    x1, y1 = flatten_point1
    x2, y2 = flatten_point2
    return ((((x2 - x1 )**2) + ((y2-y1)**2) )**0.5)

# def get_flatten_point(reference_point, point):
#     flatten_point = flatten(
#         s2sphere.LatLng.from_degrees(*reference_point),
#         s2sphere.LatLng.from_degrees(*point))

def get_flight_state_vertex(point1, point2, flight_distance, points, flight_state_vertices=[]):
    """
    Args:
        point1: starting vertex
        point2: next vertex
        flight_distance: distance covered by a flight in an interval.
        points: Iterator over all the flattened points.
        flight_state_vertices: flight state vertices collected over the flight m/s interval.
    """
    v_distance = get_distance_between_two_points(point1, point2)
    if v_distance == flight_distance:
        flight_state_vertices.append(point2)
        point1 = point2
        point2 = next(points, None)
        if not point2:
            return
        get_flight_state_vertex(
            point1, point2, flight_distance, points, flight_state_vertices)
    if v_distance < flight_distance:
        remaining_flight_distance = flight_distance - v_distance
        point1 = point2
        point2 = next(points, None)
        if not point2:
            return
        get_flight_state_vertex(
            point1, point2, remaining_flight_distance, points, flight_state_vertices)
    if flight_distance < v_distance:
        state_vertex = get_vertex_between_points(point1, point2, flight_distance)
        # As state_vertex is a Point object, get x,y coord out of it.
        if state_vertex:
            state_vertex = state_vertex.coords[:][0]
        else:
            # ?
            pass
        flight_state_vertices.append(state_vertex)
        point1 = state_vertex
        point2 = next(points, None)
        if not point2:
            return
        get_flight_state_vertex(
            point1, point2, flight_distance, points, flight_state_vertices)

def get_vertex_between_points(point1, point2, at_distance):
    """
    Args:
        point1: First vertex having tuple (x,y) co-ordinates.
        point2: Second vertex having tuple (x,y) co-ordinates.
        at_distance: A distance at which to locate the vertex on the line joining point1 and point2.
    """
    line = LineString([point1, point2])
    new_point = line.interpolate(at_distance)
    print(new_point)
    return new_point

def main():
    coordinates = get_flight_coordinates()
    reference_point = coordinates[0] # TODO: Check `if coordinates:`.
    # alt = reference_point[2] # TODO: Get alt from coordinates.
    alt = '140'
    flatten_points = []
    for point in coordinates:
        flatten_points.append(flatten(
            s2sphere.LatLng.from_degrees(*reference_point[:2]),
            s2sphere.LatLng.from_degrees(*point[:2])
        ))

    points = iter(flatten_points)
    point1 = next(points)
    point2 = next(points)
    flight_distance = get_distance_travelled()
    flight_state_vertices = []
    get_flight_state_vertex(
        point1, point2, flight_distance, points, flight_state_vertices=flight_state_vertices)
    print(f'flight_state_vertices: {flight_state_vertices}')

    flight_state_vertices_unflatten = [unflatten(s2sphere.LatLng.from_degrees(*reference_point[:2]), v) for v in flight_state_vertices]
    # flight_state_vertices_unflatten = [(p.lat(), p.lng()) for p in flight_state_vertices_unflatten]
    flight_state_vertices_unflatten = [str(p).lstrip('LatLng: ')+f',{alt}' for p in flight_state_vertices_unflatten]
    flight_state_vertices_str = '\n'.join(flight_state_vertices_unflatten)
    print(flight_state_vertices_str)
    # print(f'\n\nflight_state_vertices_unflatten: {flight_state_vertices_unflatten}, {type(flight_state_vertices_unflatten[0])}')


if __name__ == '__main__':
    main()