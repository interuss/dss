import os
from typing import Iterable, Dict

import marko.block
import marko.element
import marko.inline
from marko.md_renderer import MarkdownRenderer

from monitoring.uss_qualifier.documentation import text_of
from monitoring.uss_qualifier.requirements.documentation import RequirementID
from monitoring.uss_qualifier.scenarios.documentation.parsing import (
    get_documentation_filename,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenarioType


def format_scenario_documentation(
    test_scenarios: Iterable[TestScenarioType],
) -> Dict[str, str]:
    """Get new documentation content after autoformatting Scenario docs.

    Returns:
        Mapping from .md filename to content that file should contain (which is
        different from what it currently contains).
    """
    new_versions: Dict[str, str] = {}
    for test_scenario in test_scenarios:
        md = marko.Markdown(renderer=MarkdownRenderer)
        # Load the .md file if it exists
        doc_filename = get_documentation_filename(test_scenario)
        if not os.path.exists(doc_filename):
            continue
        with open(doc_filename, "r") as f:
            original = f.read()
        doc = md.parse(original)
        original = md.render(doc)

        _add_requirement_links(doc, doc_filename)

        # Return the formatted version
        formatted = md.render(doc)
        if formatted != original:
            new_versions[doc_filename] = formatted
    return new_versions


def _add_requirement_links(parent: marko.element.Element, doc_path: str) -> None:
    if hasattr(parent, "children") and not isinstance(parent.children, str):
        for i, child in enumerate(parent.children):
            if isinstance(child, str):
                continue
            if isinstance(child, marko.inline.StrongEmphasis):
                if not child.children:
                    raise ValueError(
                        "No content found in bolded (**strong emphasis**) section of documentation"
                    )
                if len(child.children) != 1:
                    content_types = ", ".join(
                        c.__class__.__name__ for c in child.children
                    )
                    raise ValueError(
                        f"Expected exactly 1 content element in bolded (**strong emphasis**) section of documentation, but instead found {len(child.children)} content elements ({content_types})"
                    )
                if isinstance(child.children[0], marko.inline.Link):
                    # Requirement already has link; ensure validity
                    req_id = RequirementID(text_of(child.children[0]))
                    if not os.path.exists(req_id.md_file_path()):
                        raise ValueError(
                            f'Requirement ID "{req_id}" implies that {req_id.md_file_path()} should exist, but it does not exist'
                        )
                    href = child.children[0].dest
                    doc_dir = os.path.dirname(doc_path)
                    linked_path = os.path.normpath(os.path.join(doc_dir, href))
                    if linked_path != req_id.md_file_path():
                        href = os.path.relpath(req_id.md_file_path(), doc_dir)
                        child.children[0].dest = href

                elif isinstance(child.children[0], marko.inline.RawText):
                    # Replace plaintext with link to requirement definition
                    req_id = RequirementID(text_of(child.children[0]))
                    if not os.path.exists(req_id.md_file_path()):
                        raise ValueError(
                            f'Requirement ID "{req_id}" implies that {req_id.md_file_path()} should exist, but it does not exist'
                        )
                    href = os.path.relpath(
                        req_id.md_file_path(), os.path.dirname(doc_path)
                    )
                    del child.children[0]
                    link = marko.parse(f"[{req_id}]({href})").children[0].children[0]
                    child.children.append(link)
                else:
                    raise ValueError(
                        f"Found a {child.children[0].__class__.__name__} content element in a bolded (**strong emphasis**) section of documentation, but expected either a Link or RawText"
                    )
            else:
                _add_requirement_links(child, doc_path)
