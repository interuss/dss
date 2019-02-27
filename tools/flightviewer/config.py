"""Configuration options for flightviewer.

Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""

from optparse import OptionParser
import os


def ParseOptions(argv):
  """Parses desired options from the command line and starts operations.

  Uses the command line parameters as argv, which can be altered as needed for
  testing.

  Args:
    argv: Command line parameters
  Returns:
    Options structure
  """
  parser = OptionParser(usage='usage: %prog [options]')
  parser.add_option(
    '--server',
    dest='server',
    default=os.getenv('FLIGHTVIEWER_SERVER', '127.0.0.1'),
    help='Specific server name to use on this machine',
    metavar='SERVER')
  parser.add_option(
    '--port',
    dest='port',
    default=os.getenv('FLIGHTVIEWER_PORT', '5000'),
    help='Specific port to use on this machine',
    metavar='PORT')
  parser.add_option(
    '--nodeurl',
    dest='nodeurl',
    default=os.getenv('FLIGHTVIEWER_NODEURL',
                      'https://node1.staging.interussplatform.com:8121'),
    help='Base URL of InterUSS Platform data node',
    metavar='URL')
  parser.add_option(
    '--authurl',
    dest='authurl',
    default=os.getenv('FLIGHTVIEWER_AUTHURL',
                      'https://auth.staging.interussplatform.com:8121/oauth/token?grant_type=client_credentials'),
    help='URL at which to retrieve access tokens',
    metavar='URL')
  parser.add_option(
    '-u', '--username',
    dest='username',
    default=os.getenv('FLIGHTVIEWER_USERNAME', 'flightviewer.test'),
    help='Username for retrieving access tokens',
    metavar='USERNAME')
  parser.add_option(
    '-p', '--password',
    dest='password',
    default=os.getenv('FLIGHTVIEWER_PASSWORD', ''),
    help='Password for retrieving access tokens',
    metavar='PASSWORD')
  parser.add_option(
    '--zoom',
    dest='zoom',
    default=os.getenv('FLIGHTVIEWER_ZOOM', '14'),
    help='InterUSS Platform zoom level',
    metavar='ZOOM')
  (options, args) = parser.parse_args(argv)
  del args

  return options
