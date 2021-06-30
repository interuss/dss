import s2sphere

from monitoring.monitorlib import geo


def test_flatten_unflatten():
  pts = [(34, -118), (-70, -150), (45, 9), (-10, 80), (0, 0), (1, 1), (-1, -1)]
  deltas = [(0, 0), (1e-6, 1e-6), (1e-3, 1e-3), (-1e-2, 1e-2), (1e-4, -1e-4)]

  for pt in pts:
    ref = s2sphere.LatLng.from_degrees(pt[0], pt[1])
    for delta in deltas:
      p = s2sphere.LatLng.from_degrees(pt[0] + delta[0], pt[1] + delta[1])
      xy = geo.flatten(ref, p)
      p1 = geo.unflatten(ref, xy)
      assert abs(p.lat().degrees - p1.lat().degrees) < 1e-9
      assert abs(p.lng().degrees - p1.lng().degrees) < 1e-9
