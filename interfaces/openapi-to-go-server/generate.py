# This tool generates Go server code from an OpenAPI YAML file.

import argparse
import os
from typing import List, Dict

import yaml

import apis
import formatting
import rendering


def _parse_args():
    parser = argparse.ArgumentParser(description='Preprocess an OpenAPI YAML')

    # Input/output specifications
    parser.add_argument('--api', dest='apis', type=str, action='append',
                        help='Source YAML to preprocess along with tags (if applicable) and the name of the API.  Form of --api PATH_TO_YAML#TAG1,TAG2@API_NAME')
    parser.add_argument('--api_folder', dest='api_folder', type=str,
                        default=None,
                        help='Folder that will hold the generated output for APIs')
    parser.add_argument('--example_folder', dest='example_folder', type=str,
                        default=None,
                        help='Folder that will hold the generated output for an example entrypoint and implementations')

    # General generation options
    parser.add_argument('--api_import', dest='api_import', type=str,
                        help='Full Go import path for the API package')

    return parser.parse_args()


def _generate_apis(api_list: List[apis.API], apis_folder: str, api_import: str, ensure_500: bool):
    """Generate Go libraries for APIs.

    :param api_list: APIs implemented and hosted in example
    :param apis_folder: Root location where generated Go API packages should be written
    :param api_import: Go import path for the root api package
    :param ensure_500: True to auto-generate a 500 response for each operation when one is not already declared in the API
    """
    api_package = formatting.package_of_import(api_import)
    template_vars = {
        '<IMPORTS>': rendering.imports([api_import]),
        '<API_PACKAGE>': api_package,
    }

    # Generate Go utilities common to any API generated with this tool
    os.makedirs(apis_folder, exist_ok=True)
    common_package = os.path.split(apis_folder)[-1]
    template_vars['<PACKAGE>'] = common_package
    with open(os.path.join(apis_folder, 'common.gen.go'), 'w') as f:
        f.write(rendering.template_content('header', template_vars))
        f.write(rendering.template_content('common', template_vars))
        f.write('\n')

    # Generate a package for each API
    for api in api_list:
        template_vars['<PACKAGE>'] = api.package
        api_folder = os.path.join(apis_folder, api.package)
        os.makedirs(api_folder, exist_ok=True)

        # Generate Go type definitions
        with open(os.path.join(api_folder, 'types.gen.go'), 'w') as f:
            f.write(rendering.template_content('header', template_vars))
            for data_type in api.data_types:
                f.write('\n'.join(rendering.data_type(data_type)) + '\n'*2)

        # Generate Go handler implementation interface
        template_vars['<INTERFACES>'] = '\n'.join(rendering.implementation_interface(api, api_package, ensure_500))
        with open(os.path.join(api_folder, 'interface.gen.go'), 'w') as f:
            f.write(rendering.template_content('header', template_vars))
            f.write(rendering.template_content('interface', template_vars))

        # Generate Go server factory
        template_vars['<ROUTES>'] = '\n'.join(rendering.routes(api, api_package, ensure_500))
        template_vars['<ROUTING>'] = '\n'.join(rendering.routing(api, api_package))
        with open(os.path.join(api_folder, 'server.gen.go'), 'w') as f:
            f.write(rendering.template_content('header', template_vars))
            f.write(rendering.template_content('server', template_vars))


def _generate_example(api_list: List[apis.API], output_folder: str, api_import: str):
    """Generate example implementations and entry point.

    :param api_list: APIs implemented and hosted in example
    :param output_folder: Location where example Go code should be written
    :param api_import: Go import path for the root api package
    """
    api_package = formatting.package_of_import(api_import)

    # Generate example Go API implementations
    implementations: Dict[str, str] = {}
    implementation_lines: List[str] = []
    for api in api_list:
        implementation = formatting.snake_case_to_pascal_case(api.package) + 'Implementation'
        implementation_lines.extend(rendering.example_implementation(api, implementation) + [''])
        implementations[api.package] = implementation

    # Generate Go router definitions
    router_def_lines = rendering.example_router_defs(implementations, api_package)

    # Write main.gen.go using main.go.template
    imports = [api_import + '/' + api.package for api in api_list] + [api_import]
    template_vars = {
        '<PACKAGE>': 'main',
        '<IMPORTS>': rendering.imports(imports),
        '<API_PACKAGE>': formatting.package_of_import(api_import),
        '<IMPLEMENTATIONS>': '\n'.join(implementation_lines),
        '<ROUTER_DEFS>': '\n'.join(router_def_lines),
    }
    with open(os.path.join(output_folder, 'main.gen.go'), 'w') as f:
        f.write(rendering.template_content('header', template_vars))
        f.write(rendering.template_content('main', template_vars))


def main():
    args = _parse_args()

    # Parse API definitions
    api_list: List[apis.API] = []
    for api_declaration in args.apis:
        if '@' in api_declaration:
            input_yaml, package = api_declaration.split('@')
            api_path = package
        else:
            input_yaml = api_declaration
            package = os.path.split(api_declaration)[-1].split('.')[0]
            api_path = ''
        package = package.replace('-', '_')

        if '#' in input_yaml:
            input_yaml, tag_list = input_yaml.split('#')
            tags = {t.strip() for t in tag_list.split(',')}
        else:
            tags = set()

        with open(input_yaml, mode='r') as f:
            spec = yaml.full_load(f)

        api = apis.make_api(package, api_path, spec)
        if tags:
            api.filter_operations(tags)
        api_list.append(api)

    # Render Go code
    if args.api_folder:
        _generate_apis(api_list, args.api_folder, args.api_import, True)
        os.system('cd {} && gofmt -s -w .'.format(args.api_folder))
    if args.example_folder:
        _generate_example(api_list, args.example_folder, args.api_import)
        os.system('cd {} && gofmt -s -w .'.format(args.example_folder))


if __name__ == '__main__':
    main()
