import sys

from monitoring.rid_qualifier.host import webapp


def main(argv):
  del argv
  webapp.run(host='localhost', port=8072)


if __name__ == '__main__':
  main(sys.argv)
