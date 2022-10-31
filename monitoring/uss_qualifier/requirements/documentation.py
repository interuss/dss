import os
from typing import Dict, Set

from implicitdict import ImplicitDict
import marko
import marko.element
import marko.inline


RequirementID = str
"""Identifier for a requirement.

Form: <PACKAGE>.<NAME>

PACKAGE is a Python-style package reference to a .md file (without extension)
relative to uss_qualifier/requirements.  For instance, the PACKAGE for the file
located at uss_qualifier/requirements/astm/f3548/v21.md would be
`astm.f3548.v21`.

NAME is an identifier defined in the file described by PACKAGE by enclosing it
in a <tt> tag; for instance: `<tt>USS0105</tt>`.
"""


class Requirement(object):
    def __init__(self, requirement_id: RequirementID):
        self.requirement_id = requirement_id


class RequirementSet(ImplicitDict):
    name: str
    requirement_ids: RequirementID


_verified_requirements: Set[RequirementID] = set()


def _text_of(parent: marko.element.Element) -> str:
    if not hasattr(parent, "children"):
        return ""
    if isinstance(parent.children, str):
        return parent.children
    return "".join(_text_of(c) for c in parent.children)


def _verify_requirements(parent: marko.element.Element, package: str) -> None:
    if hasattr(parent, "children") and not isinstance(parent.children, str):
        for i, child in enumerate(parent.children):
            if isinstance(child, str):
                continue
            if (
                i < len(parent.children) - 2
                and isinstance(child, marko.inline.InlineHTML)
                and child.children == "<tt>"
                and isinstance(parent.children[i + 2], marko.inline.InlineHTML)
                and parent.children[i + 2].children == "</tt>"
            ):
                name = _text_of(parent.children[i + 1])
                _verified_requirements.add(package + "." + name)
            else:
                _verify_requirements(child, package)


def _load_requirement(requirement_id: RequirementID) -> None:
    illegal_characters = "#%&{}\\<>*?/ $!'\":@+`|="
    if any(c in requirement_id for c in illegal_characters):
        raise ValueError(
            f'Requirement ID "{requirement_id}" may not contain any of these characters: {illegal_characters}'
        )
    parts = requirement_id.split(".")
    name = parts[-1]
    package = ".".join(parts[:-1])
    md_filename = os.path.abspath(
        os.path.join(os.path.dirname(__file__), os.path.join(*parts[0:-1]) + ".md")
    )
    if not os.path.exists(md_filename):
        raise ValueError(
            f'Could not load requirement "{requirement_id}" because the file "{md_filename}" does not exist'
        )
    with open(md_filename, "r") as f:
        doc = marko.parse(f.read())
    _verify_requirements(doc, package)
    if requirement_id not in _verified_requirements:
        raise ValueError(
            f'Requirement "{name}" could not be found in "{md_filename}", so the requirement {requirement_id} could not be loaded (the file must contain `<tt>{name}</tt>` somewhere in it, but does not)'
        )


def get_requirement(requirement_id: RequirementID) -> Requirement:
    if requirement_id not in _verified_requirements:
        _load_requirement(requirement_id)
    return Requirement(requirement_id)
