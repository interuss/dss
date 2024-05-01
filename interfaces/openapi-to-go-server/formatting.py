from typing import Dict
import re

def capitalize_first_letter(s: str) -> str:
    return s[0].upper() + s[1:] if s else s


def snake_case_to_pascal_case(s: str) -> str:
    return s.replace('_', ' ').title().replace(' ', '')


def string_to_pascal_case(s: str) -> str:
    return re.sub(r'[^a-zA-Z0-9]', ' ', s).title().replace(' ', '')


def replace(s: str, template_vars: Dict[str, str]) -> str:
    for k, v in template_vars.items():
        s = s.replace(k, v)
    return s


def package_of_import(import_path: str) -> str:
    return import_path.split('/')[-1]
