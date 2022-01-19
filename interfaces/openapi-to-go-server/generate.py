# This tool generates Go server code from an OpenAPI YAML file.

import argparse
import os

import yaml

import data_types
import operations
import rendering


def _parse_args():
    parser = argparse.ArgumentParser(description='Preprocess an OpenAPI YAML')

    # Input/output specifications
    parser.add_argument('--input_yaml', dest='input_yaml', type=str,
                        help='source YAML to preprocess')
    parser.add_argument('--output_folder', dest='output_folder', type=str,
                        default='.',
                        help='destination filename to write resulting YAML')

    # General generation options
    parser.add_argument('--include_endpoint_tags', dest='include_endpoint_tags',
                        type=str, default='',
                        help='comma-separated list of tags for which to include endpoints')
    parser.add_argument('--package', dest='package', type=str,
                        default='openapi', help='Go package name for the output')
    parser.add_argument('--path_prefix', dest='path_prefix', type=str,
                        default='',
                        help='Prefix to prepend to all paths when generating routers')

    # Options indicating which components to generate
    parser.add_argument('--include_types', dest='include_types',
                        default=False, action='store_true',
                        help='Generate data type definitions types.gen.go in the output folder')
    parser.add_argument('--include_interface', dest='include_interface',
                        default=False, action='store_true',
                        help='Generate implementation interface interface.gen.go in the output folder')
    parser.add_argument('--include_server', dest='include_server',
                        default=False, action='store_true',
                        help='Generate HTTP server definition server.gen.go in the output folder')
    parser.add_argument('--include_common', dest='include_common',
                        default=False, action='store_true',
                        help='Generate common.gen.go in the output folder')
    parser.add_argument('--include_example', dest='include_example',
                        default=False, action='store_true',
                        help='Generate example implementation (main.gen.go)')

    return parser.parse_args()


def main():
    args = _parse_args()
    with open(args.input_yaml, mode='r') as f:
        spec = yaml.full_load(f)

    if args.include_types:
        # Parse all defined data types
        if 'components' not in spec:
            raise ValueError('Missing `components` in YAML')
        components = spec['components']
        if 'schemas' not in components:
            raise ValueError('Missing `schemas` in `components`')
        declared_types = []
        for name, schema in components['schemas'].items():
            data_type, additional_types = data_types.make_data_types(name, schema)
            declared_types.extend(additional_types)
            declared_types.append(data_type)

        # Generate Go type definitions
        with open(os.path.join(args.output_folder, 'types.gen.go'), 'w') as f:
            f.write('\n'.join(rendering.header(args.package)) + '\n'*2)
            for data_type in declared_types:
                f.write('\n'.join(rendering.data_type(data_type)) + '\n'*2)

    # Parse all endpoints
    if 'paths' not in spec:
        raise ValueError('Missing `paths` in YAML')
    paths = spec['paths']
    declared_operations = []
    for name, schema in paths.items():
        new_endpoints = operations.make_operations(name, schema)
        declared_operations.extend(new_endpoints)

    # Filter endpoints by tags, if specified
    if args.include_endpoint_tags:
        tags = args.include_endpoint_tags.split(',')
        declared_operations = [endpoint for endpoint in declared_operations
                              if not all(tag not in endpoint.tags for tag in tags)]

    # Generate Go handler implementation interface
    if args.include_interface:
        with open(os.path.join(args.output_folder, 'interface.gen.go'), 'w') as f:
            f.write('\n'.join(rendering.header(args.package)) + '\n'*2)
            f.write('\n'.join(rendering.implementation_interface(declared_operations)))
            f.write('\n')

    # Generate Go server factory
    if args.include_server:
        with open(os.path.join(args.output_folder, 'server.gen.go'), 'w') as f:
            f.write('\n'.join(rendering.header(args.package)) + '\n'*2)
            with open('templates/server.go.template', 'r') as t:
                f.write(t.read())
            f.write('\n')
            f.write('\n'.join(rendering.routes(declared_operations, args.path_prefix)))
            f.write('\n')

    # Generate Go main (executable) example
    if args.include_example:
        with open(os.path.join(args.output_folder, 'main.gen.go'), 'w') as f:
            f.write('\n'.join(rendering.header(args.package)) + '\n'*2)
            with open('templates/main.go.template', 'r') as t:
                f.write(t.read())
            f.write('\n')
            f.write('\n'.join(rendering.example(declared_operations)))
            f.write('\n')

    # Generate Go utilities common to any API generated with this tool
    if args.include_common:
        with open(os.path.join(args.output_folder, 'common.gen.go'), 'w') as f:
            f.write('\n'.join(rendering.header(args.package)) + '\n'*2)
            with open('templates/common.go.template', 'r') as t:
                f.write(t.read())
            f.write('\n')

    # Tidy up all the formatting
    os.system('cd {} && gofmt -s -w .'.format(args.output_folder))


if __name__ == '__main__':
    main()
