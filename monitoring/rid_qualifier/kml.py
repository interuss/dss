# Module to parse KML file.

import logging
from pykml import parser
import re


KML_NAMESPACE = {"kml":"http://www.opengis.net/kml/2.2"}

def get_kml_root(kml_obj, from_string=False):
    print('kml_obj: ', type(kml_obj))
    logging.info(f'kml_obj: {type(kml_obj)}')
    if from_string:
        content = parser.fromstring(kml_obj)
        return content
    content = parser.parse(kml_obj)
    return content.getroot()


def get_folders(root):
    return root.Document.Folder.Folder


def get_polygon_speed(polygon_name):
    """Returns speed unit within a polygon."""
    result = re.search(r"\(([0-9.]+)\)", polygon_name)
    return float(result.group(1)) if result else None

def get_folder_details(folder_elem):
    speed_polygons = {}
    alt_polygons = {}
    operator_location = {}
    coordinates = ''
    for placemark in folder_elem.xpath('.//kml:Placemark', namespaces=KML_NAMESPACE):
        placemark_name = str(placemark.name)
        polygons = placemark.xpath('.//kml:Polygon', namespaces=KML_NAMESPACE)

        if placemark_name == 'operator_location':
            operator_point = folder_elem.xpath(
                    './/kml:Placemark/kml:Point/kml:coordinates', namespaces=KML_NAMESPACE)[0]
            if operator_point:
                operator_point = str(operator_point).split(',')
                operator_location = {
                    'lng': operator_point[0],
                    'lat': operator_point[1]
                }
        if polygons:
            if placemark_name.startswith('alt:'):
                polygon_coords = get_coordinates_from_kml(
                    polygons[0].outerBoundaryIs.LinearRing.coordinates)
                alt_polygons.update({placemark_name: polygon_coords})
            if placemark_name.startswith('speed:'):
                if not get_polygon_speed(placemark_name):
                    raise ValueError('Could not determine Polygon speed from Placemark "{}"'.format(placemark_name))
                polygon_coords = get_coordinates_from_kml(
                    polygons[0].outerBoundaryIs.LinearRing.coordinates)
                speed_polygons.update({placemark_name: polygon_coords})
        
        coords = placemark.xpath('.//kml:LineString/kml:coordinates', namespaces=KML_NAMESPACE)
        if coords:
            coordinates = coords
            coordinates = get_coordinates_from_kml(coordinates)
    return  {
        str(folder_elem.name): {
            'description': get_folder_description(folder_elem),
            'speed_polygons': speed_polygons,
            'alt_polygons': alt_polygons,
            'input_coordinates': coordinates,
            'operator_location': operator_location
    }}


def get_coordinates_from_kml(coordinates):
    """Returns list of tuples of coordinates.
    Args:
        coordinates: coordinates element from KML.
    """
    if coordinates:
       return [tuple(float(x.strip()) for x in c.split(',')) for c in str(coordinates[0]).split(' ') if c.strip()]


def get_folder_description(folder_elem):
    """Returns folder description from KML.
    Args:
        folder_elem: Folder element from KML.
    """
    description = folder_elem.description
    return dict([tuple(j.strip() for j in i.split(':')) for i in str(description).split('\n')])


def get_kml_content(kml_file, from_string=False):
    root = get_kml_root(kml_file, from_string)
    folders = get_folders(root)
    kml_content = {}
    for folder in folders:
        folder_details = get_folder_details(folder)
        if folder_details:
            kml_content.update(folder_details)
    return kml_content
