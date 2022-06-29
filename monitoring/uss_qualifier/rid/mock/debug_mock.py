#!env/bin/python3

from monitoring.uss_qualifier.rid.mock import webapp


if __name__ == "__main__":
    webapp.run("localhost", 8070)
