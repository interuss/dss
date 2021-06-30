# Module to parse KML file.

from pykml import parser
import re


KML_NAMESPACE = {"kml":"http://www.opengis.net/kml/2.2"}

def get_kml_root(kml_path):
    kml = parser.parse(kml_path)
    return kml.getroot()


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
            operator_location = {
                'lng': folder_elem.xpath(
                    './/kml:Placemark/kml:LookAt/kml:longitude', namespaces=KML_NAMESPACE)[0],
                'lat': folder_elem.xpath(
                    './/kml:Placemark/kml:LookAt/kml:latitude', namespaces=KML_NAMESPACE)[0]
            }
        if polygons:
            if placemark_name.startswith('alt:'):
                alt_polygons.update({placemark.name: polygons[0].outerBoundaryIs.LinearRing.coordinates})
            if placemark_name.startswith('speed:'):
                if not get_polygon_speed(placemark_name):
                    # TODO: raise error
                    pass
                speed_polygons.update({placemark.name: polygons[0].outerBoundaryIs.LinearRing.coordinates})
        
        coords = placemark.xpath('.//kml:LineString/kml:coordinates', namespaces=KML_NAMESPACE)
        if coords:
            coordinates = coords
            coordinates = get_coordinates_from_kml(coordinates)
        print(f'from xpath placemarks: {placemark.name}....{len(polygons)}')
    
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

def get_folder_description(folder_elem, to_json=True):
    """Returns folder description from KML.
    Args:
        folder_elem: Folder element from KML.
        to_json: If True: Returns description details as key: value pair, else a description string.
    """
    if to_json:
        description = folder_elem.description
        return dict([tuple((i.strip()).split(':')) for i in str(description).split('\n')])
    return str(folder_elem.description)


def get_kml_content(kml_file):
    root = get_kml_root(kml_file)
    folders = get_folders(root)
    kml_content = {}
    for folder in folders:
        print(folder.name)
        folder_details = get_folder_details(folder)
        if folder_details:
            kml_content.update(folder_details)
    return kml_content

if __name__ == '__main__':
    kml_path = 'monitoring/rid_qualifier/test_data/dcdemo.kml'
    kml_content = get_kml_content(kml_path)
    print(kml_content)