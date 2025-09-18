# Small python script to do some evict tests. Use only standard libraries to
# run everywhere.
# Expect that `start-locally` have been ran

import sys
import logging
import os
import importlib

from evict_helper import EvictHelper
from query_helper import QueryHelper


logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s %(levelname)-8s %(name)-24s %(message)-50s",
    handlers=[logging.StreamHandler(sys.stdout)],
)

qh = QueryHelper()
eh = EvictHelper()

dir_path = os.path.dirname(os.path.realpath(__file__))
tests_path = os.path.join(dir_path, "tests")

for f in sorted(os.listdir(tests_path)):
    if f != "__init__.py" and f.endswith(".py"):
        name = f[:-3]

        module = importlib.import_module(f"tests.{name}")
        getattr(module, name)(qh, eh)
