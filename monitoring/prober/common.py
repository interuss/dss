DATE_FORMAT = '%Y-%m-%dT%H:%M:%SZ'

VERTICES = [
    {'lat': 130.6205, 'lng': -23.6558},
    {'lat': 130.6301, 'lng': -23.6898},
    {'lat': 130.6700, 'lng': -23.6709},
    {'lat': 130.6466, 'lng': -23.6407},
]

GEO_POLYGON_STRING = ','.join('{},{}'.format(x['lat'], x['lng']) for x in VERTICES)
