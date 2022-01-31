import sys

from monitoring.mock_uss import webapp


def main(argv):
  del argv
  webapp.run(host='localhost', port=8071)


if __name__ == '__main__':
  main(sys.argv)
