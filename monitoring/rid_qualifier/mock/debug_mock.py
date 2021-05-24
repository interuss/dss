#!env/bin/python3

from monitoring.rid_qualifier.mock import webapp


if __name__ == "__main__":
  webapp.run('localhost', 8070)
