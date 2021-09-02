import sys

from monitoring.mock_riddp import webapp


def main(argv):
  del argv
  webapp.run(host='localhost', port=8073)


if __name__ == '__main__':
  main(sys.argv)
