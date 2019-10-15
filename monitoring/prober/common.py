MAX_SUB_PER_AREA = 10

DATE_FORMAT = '%Y-%m-%dT%H:%M:%SZ'

VERTICES = [
    {
        'lat': 130.6205,
        'lng': -23.6558
    },
    {
        'lat': 130.6301,
        'lng': -23.6898
    },
    {
        'lat': 130.6700,
        'lng': -23.6709
    },
    {
        'lat': 130.6466,
        'lng': -23.6407
    },
]

GEO_POLYGON_STRING = ','.join(
    '{},{}'.format(x['lat'], x['lng']) for x in VERTICES)

HUGE_VERTICES = [
    {
        'lat': 130,
        'lng': -23
    },
    {
        'lat': 130,
        'lng': -24
    },
    {
        'lat': 132,
        'lng': -24
    },
    {
        'lat': 132,
        'lng': -23
    },
]

HUGE_GEO_POLYGON_STRING = ','.join(
    '{},{}'.format(x['lat'], x['lng']) for x in HUGE_VERTICES)
