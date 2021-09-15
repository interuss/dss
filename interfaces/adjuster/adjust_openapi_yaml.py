# Our OpenAPI interface -> openapi2proto -> grpc-gateway -> grpc toolchain
# does not currently deal correctly with a number of OpenAPI features.
# This tool rewrites an OpenAPI YAML file to get openapi2proto to produce a
# proto that will correctly parse properly-formatted input and produce
# properly-formatted output after going through the gateway layer.

import argparse

import yaml


# Whenever node or a descendant contains an "enum" key and "string" in its
# "type" key, replace the "enum" key with an "example" key containing the first
# value in the original "enum" value list
def fix_string_enums(node):
  if isinstance(node, dict) and 'enum' in node and node.get('type', None) == 'string':
    node['example'] = node['enum'][0]
    del node['enum']
  for _, v in node.items():
    if isinstance(v, dict):
      fix_string_enums(v)


# Specific to the SCD API, manually change the type of the `key` field to array
# of string
def fix_key_type(tree):
  keys = [
    tree['components']['schemas']['PutOperationalIntentReferenceParameters']['properties']['key'],
    tree['components']['schemas']['GeoZone']['properties']['identifier'],
    tree['components']['schemas']['GeoZone']['properties']['country'],
    tree['components']['schemas']['GeoZone']['properties']['name'],
    tree['components']['schemas']['GeoZone']['properties']['type'],
    tree['components']['schemas']['GeoZone']['properties']['restriction'],
    tree['components']['schemas']['GeoZone']['properties']['regulation_exemption'],
    tree['components']['schemas']['GeoZone']['properties']['u_space_class'],
    tree['components']['schemas']['GeoZone']['properties']['message'],
  ]

  for key in keys:
    del key['anyOf']
    key['type'] = 'array'
    key['items'] = {
      'type': 'string'
    }


parser = argparse.ArgumentParser(description='Preprocess an OpenAPI YAML')
parser.add_argument('--input_yaml', dest='input_yaml', type=str,
                    help='source YAML to preprocess')
parser.add_argument('--output_yaml', dest='output_yaml', type=str,
                    help='destination filename to write resulting YAML')
args = parser.parse_args()

with open(args.input_yaml, mode='r') as f:
  spec = yaml.full_load(f)

fix_string_enums(spec)
fix_key_type(spec)

print('Writing to ' + args.output_yaml)
with open(args.output_yaml, mode='w') as f:
  yaml.dump(spec, f)
