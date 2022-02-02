from typing import Dict


def capitalize_first_letter(s: str) -> str:
    return s[0].upper() + s[1:] if s else s


def snake_case_to_pascal_case(s: str) -> str:
    return s.replace('_', ' ').title().replace(' ', '')


def replace(s: str, template_vars: Dict[str, str]) -> str:
    for k, v in template_vars.items():
        s = s.replace(k, v)
    return s


def package_of_import(import_path: str) -> str:
    return import_path.split('/')[-1]
