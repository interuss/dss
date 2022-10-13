import os
import sys

from monitoring.mock_uss import webapp


def main(argv):
    del argv
    port = int(os.environ.get("MOCK_USS_PORT", "8071"))
    webapp.run(host="localhost", port=port)


if __name__ == "__main__":
    main(sys.argv)
