import os
from typing import Dict, Set, List

from implicitdict import ImplicitDict
import marko
import marko.element
import marko.inline

from monitoring.uss_qualifier.documentation import text_of


class RequirementID(str):
    """Identifier for a requirement.

    Form: <PACKAGE>.<NAME>

    PACKAGE is a Python-style package reference to a .md file (without extension)
    relative to uss_qualifier/requirements.  For instance, the PACKAGE for the file
    located at uss_qualifier/requirements/astm/f3548/v21.md would be
    `astm.f3548.v21`.

    NAME is an identifier defined in the file described by PACKAGE by enclosing it
    in a <tt> tag; for instance: `<tt>USS0105</tt>`.
    """

    def __new__(cls, value):
        illegal_characters = "#%&{}\\<>*?/ $!'\":@+`|="
        if any(c in value for c in illegal_characters):
            raise ValueError(
                f'RequirementID "{value}" may not contain any of these characters: {illegal_characters}'
            )
        str_value = str.__new__(cls, value)
        return str_value

    def md_file_path(self) -> str:
        parts = self.split(".")
        md_filename = os.path.abspath(
            os.path.join(os.path.dirname(__file__), os.path.join(*parts[0:-1]) + ".md")
        )
        return md_filename

    def requirement_name(self) -> str:
        parts = self.split(".")
        return parts[-1]

    def package(self) -> str:
        parts = self.split(".")
        return ".".join(parts[:-1])


class Requirement(object):
    def __init__(self, requirement_id: RequirementID):
        self.requirement_id = requirement_id


_verified_requirements: Set[RequirementID] = set()


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
                name = text_of(parent.children[i + 1])
                _verified_requirements.add(RequirementID(package + "." + name))
            else:
                _verify_requirements(child, package)


def _load_requirement(requirement_id: RequirementID) -> None:
    md_filename = requirement_id.md_file_path()
    if not os.path.exists(md_filename):
        raise ValueError(
            f'Could not load requirement "{requirement_id}" because the file "{md_filename}" does not exist'
        )
    with open(md_filename, "r") as f:
        doc = marko.parse(f.read())
    _verify_requirements(doc, requirement_id.package())
    if requirement_id not in _verified_requirements:
        raise ValueError(
            f'Requirement "{requirement_id.requirement_name()}" could not be found in "{md_filename}", so the requirement {requirement_id} could not be loaded (the file must contain `<tt>{requirement_id.requirement_name()}</tt>` somewhere in it, but does not)'
        )


def get_requirement(requirement_id: RequirementID) -> Requirement:
    if requirement_id not in _verified_requirements:
        _load_requirement(requirement_id)
    return Requirement(requirement_id)


class RequirementSetID(str):
    """Identifier for a set of requirements.

    The form of a value is a Python-style package reference to a .md file (without
    extension) relative to uss_qualifier/requirements.  For instance, the set of
    requirements described in uss_qualifier/requirements/astm/f3548/v21/scd.md would
    have a RequirementSetID of astm.f3548.v21.scd.
    """

    def __new__(cls, value):
        illegal_characters = "#%&{}\\<>*?/ $!'\":@+`|="
        if any(c in value for c in illegal_characters):
            raise ValueError(
                f'RequirementSetID "{value}" may not contain any of these characters: {illegal_characters}'
            )
        str_value = str.__new__(cls, value)
        return str_value


class RequirementSet(ImplicitDict):
    name: str
    requirement_ids: List[RequirementID]


REQUIREMENT_SET_SUFFIX = " requirement set"


_requirement_sets: Dict[RequirementSetID, RequirementSet] = {}


def _parse_requirements(parent: marko.element.Element) -> List[RequirementID]:
    reqs = []
    if hasattr(parent, "children") and not isinstance(parent.children, str):
        for i, child in enumerate(parent.children):
            if isinstance(child, str):
                continue
            if isinstance(child, marko.inline.StrongEmphasis):
                req_id = text_of(parent.children[i])
                reqs.append(RequirementID(req_id))
            else:
                reqs.extend(_parse_requirements(child))
    return reqs


def _load_requirement_set(requirement_set_id: RequirementSetID) -> RequirementSet:
    parts = requirement_set_id.split(".")
    md_filename = os.path.abspath(
        os.path.join(os.path.dirname(__file__), os.path.join(*parts) + ".md")
    )
    if not os.path.exists(md_filename):
        raise ValueError(
            f'Could not load requirement set "{requirement_set_id}" because the file "{md_filename}" does not exist'
        )
    with open(md_filename, "r") as f:
        doc = marko.parse(f.read())

    # Extract the requirement set name from the first top-level header
    if (
        not isinstance(doc.children[0], marko.block.Heading)
        or doc.children[0].level != 1
        or not text_of(doc.children[0]).lower().endswith(REQUIREMENT_SET_SUFFIX)
    ):
        raise ValueError(
            f'The first line of {md_filename} must be a level-1 heading with the name of the scenario + "{REQUIREMENT_SET_SUFFIX}" (e.g., "# ASTM F3411-19 Service Provider requirement set")'
        )
    requirement_set_name = text_of(doc.children[0])[0 : -len(REQUIREMENT_SET_SUFFIX)]

    reqs = _parse_requirements(doc)
    for req in reqs:
        try:
            get_requirement(req)
        except ValueError as e:
            raise ValueError(
                f'Error loading requirement set "{requirement_set_id}" from {md_filename}: {str(e)}'
            )
    return RequirementSet(name=requirement_set_name, requirement_ids=reqs)


def get_requirement_set(requirement_set_id: RequirementSetID) -> RequirementSet:
    if requirement_set_id not in _requirement_sets:
        _requirement_sets[requirement_set_id] = _load_requirement_set(
            requirement_set_id
        )
    return _requirement_sets[requirement_set_id]
