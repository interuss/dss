VERTICES = [
    {
        'lng': 130.6205,
        'lat': -23.6558
    },
    {
        'lng': 130.6301,
        'lat': -23.6898
    },
    {
        'lng': 130.6700,
        'lat': -23.6709
    },
    {
        'lng': 130.6466,
        'lat': -23.6407
    },
]

GEO_POLYGON_STRING = ','.join(
    '{},{}'.format(x['lat'], x['lng']) for x in VERTICES)

HUGE_VERTICES = [
    {
        'lng': 130,
        'lat': -23
    },
    {
        'lng': 130,
        'lat': -24
    },
    {
        'lng': 132,
        'lat': -24
    },
    {
        'lng': 132,
        'lat': -23
    },
]

HUGE_GEO_POLYGON_STRING = ','.join(
    '{},{}'.format(x['lat'], x['lng']) for x in HUGE_VERTICES)
