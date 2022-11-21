import os.path

import marko
import marko.block
import marko.inline


REPO_CONTENT_BASE_URL = "https://github.com/interuss/dss/tree/master/"


def check_local_links(parent: marko.block.Element, doc_path: str, repo_root: str) -> None:
    if isinstance(parent, marko.inline.Link):
        if parent.dest.startswith(REPO_CONTENT_BASE_URL):
            relative_path = parent.dest[len(REPO_CONTENT_BASE_URL):]
        elif parent.dest.startswith("http://") or parent.dest.startswith("https://"):
            # Don't check absolute paths to other locations
            relative_path = None
        else:
            md_path = os.path.relpath(os.path.dirname(doc_path), repo_root)
            relative_path = os.path.join(md_path, parent.dest)
        if relative_path is not None:
            if "#" in relative_path:
                relative_path = relative_path.split("#")[0]
            abs_path = os.path.realpath(os.path.join(repo_root, relative_path))
            if not os.path.exists(abs_path):
                md_relative_path = os.path.relpath(doc_path, repo_root)
                raise ValueError(f"Document {md_relative_path} has a link to {parent.dest} but {relative_path} does not exist in the repository ({abs_path} in the repo_hygiene container)")
    else:
        if hasattr(parent, "children") and not isinstance(parent.children, str):
            for child in parent.children:
                check_local_links(child, doc_path, repo_root)
