# Small python script to do some evict tests. Use only standard libraries to
# run everywhere.

import sys
import logging
import os
import importlib

from cm_helper import CMHelper

logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s %(levelname)-8s %(name)-24s %(message)-50s",
    handlers=[logging.StreamHandler(sys.stdout)],
)


cm = CMHelper()


dir_path = os.path.dirname(os.path.realpath(__file__))
tests_path = os.path.join(dir_path, "tests")

for f in sorted(os.listdir(tests_path)):
    if f != "__init__.py" and f.endswith(".py"):
        name = f[:-3]

        module = importlib.import_module(f"tests.{name}")

        cm.prepare()
        getattr(module, name)(cm)
