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
    '-o', '--origin',
    dest='origin',
    default=os.getenv('DIZZY_ORIGIN', '37.6281,-122.4264'),
    help='latitude,longitude of the center of circular sim flights',
    metavar='LAT,LNG')
  parser.add_option(
    '-r',
    '--radius',
    dest='radius',
    default=os.getenv('DIZZY_RADIUS', '100'),
    help='Radius, in meters, of circular sim flights',
    metavar='RADIUS')
  parser.add_option(
    '--server',
    dest='server',
    default=os.getenv('DIZZY_SERVER', '127.0.0.1'),
    help='Specific server name to use on this machine',
    metavar='SERVER')
  parser.add_option(
    '--port',
    dest='port',
    default=os.getenv('DIZZY_PORT', '5000'),
    help='Specific port to use on this machine',
    metavar='PORT')
  parser.add_option(
    '--flightperiod',
    dest='flightperiod',
    default=os.getenv('DIZZY_FLIGHTPERIOD', '60'),
    help='Number of seconds each flight lasts',
    metavar='SECONDS')
  parser.add_option(
    '--flightinterval',
    dest='flightinterval',
    default=os.getenv('DIZZY_FLIGHTINTERVAL', '10'),
    help='Number of seconds between flights',
    metavar='SECONDS')
  parser.add_option(
    '-f', '--minaltitude',
    dest='minaltitude',
    default=os.getenv('DIZZY_MINALTITUDE', '60'),
    help='Minimum flight altitude, meters MSL',
    metavar='METERS')
  parser.add_option(
    '-c', '--maxaltitude',
    dest='maxaltitude',
    default=os.getenv('DIZZY_MAXALTITUDE', '140'),
    help='Maximum flight altitude, meters MSL',
    metavar='METERS')
  parser.add_option(
    '--hanger',
    dest='hanger',
    default=os.getenv('DIZZY_HANGER', 'hanger.json'),
    help='JSON hanger file',
    metavar='FILENAME')
  parser.add_option(
    '--nodeurl',
    dest='nodeurl',
    default=os.getenv('DIZZY_NODEURL',
                      'https://node1.staging.interussplatform.com:8121'),
    help='Base URL of InterUSS Platform data node',
    metavar='URL')
  parser.add_option(
    '--authurl',
    dest='authurl',
    default=os.getenv('DIZZY_AUTHURL',
                      'https://auth.staging.interussplatform.com:8121/oauth/token?grant_type=client_credentials'),
    help='URL at which to retrieve access tokens',
    metavar='URL')
  parser.add_option(
    '--baseurl',
    dest='baseurl',
    default=os.getenv('DIZZY_BASEURL', 'https//localhost:5000'),
    help='Base URL for public_portal_endpoint and flight_info_endpoint',
    metavar='URL')
  parser.add_option(
    '-k', '--authpublickey',
    dest='authpublickey',
    default=os.getenv('DIZZY_AUTHPUBLICKEY',
                      'https://auth.staging.interussplatform.com:8121/key'),
    help='URL at which to retrieve the public key to validate access tokens, '
         'or the public key itself',
    metavar='URL|KEY')
  parser.add_option(
    '-u', '--username',
    dest='username',
    default=os.getenv('DIZZY_USERNAME', 'dizzy.test'),
    help='Username for retrieving access tokens',
    metavar='USERNAME')
  parser.add_option(
    '-p', '--password',
    dest='password',
    default=os.getenv('DIZZY_PASSWORD', ''),
    help='Password for retrieving access tokens',
    metavar='PASSWORD')
  parser.add_option(
    '--zoom',
    dest='zoom',
    default=os.getenv('DIZZY_ZOOM', '14'),
    help='InterUSS Platform zoom level',
    metavar='ZOOM')
  (options, args) = parser.parse_args(argv)
  del args

  return options
