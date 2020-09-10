#!env/bin/python3

from monitoring.tracer.uss_receiver import webapp


if __name__ == "__main__":
  # Note: enabling debug=True here will cause Subscription management to be
  # performed multiple times in possibly-unexpected ways.
  webapp.run('localhost', 5000)
