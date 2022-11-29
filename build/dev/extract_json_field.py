"""Simple script to extract a field from a JSON representation on stdin."""

import json
import sys

s = ""
try:
    for line in sys.stdin:
        s += line + "\n"
except KeyboardInterrupt:
    pass

try:
    obj = json.loads(s)
except ValueError as e:
    raise ValueError(str(e) + "\nvvvvv Unable to decode JSON: vvvvv\n" + s + "^^^^^ Unable to decode JSON: ^^^^^")
fields = sys.argv[1].split(".")
for field in fields:
    obj = obj[field]
print(obj)
