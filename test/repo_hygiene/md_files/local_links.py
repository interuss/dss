import os.path

import marko
import marko.block
import marko.inline

REPO_CONTENT_BASE_URL = "https://github.com/interuss/dss/tree/master/"
DOCS_BASE_URL = "https://interuss.github.io/dss/dev/"


def _abs_path(repo_root: str, relative_path: str) -> str:
    return os.path.realpath(os.path.join(repo_root, relative_path))


def check_local_links(parent: marko.block.Element, doc_path: str, repo_root: str) -> None:
    def check(rel_paths: str | list[str]) -> None:
        """Raises an exception if all provided relative paths do not exist."""
        if isinstance(rel_paths, str):
            rel_paths = [rel_paths]

        abs_paths = [_abs_path(repo_root, rel_path) for rel_path in rel_paths]
        if all([not os.path.exists(abs_path) for abs_path in abs_paths]):
            md_relative_path = os.path.relpath(doc_path, repo_root)
            raise ValueError(f"Document {md_relative_path} has a link to {parent.dest} but none of {','.join(rel_paths)} exist in the repository ({','.join(abs_paths)} in the repo_hygiene container)")

    if isinstance(parent, marko.inline.Link):
        relative_path = parent.dest
        if "#" in relative_path:
            relative_path = relative_path.split("#")[0]

        if relative_path.startswith(REPO_CONTENT_BASE_URL):
            check(relative_path[len(REPO_CONTENT_BASE_URL):])
        elif relative_path.startswith(DOCS_BASE_URL):
            # Check links to documentation hosted on interuss.github.io/dss
            relative_path = relative_path[len(DOCS_BASE_URL):]
            if relative_path.endswith("/"):
                relative_path = relative_path[:-1]
            if relative_path.endswith("/index.html"):
                relative_path = relative_path[:-11]
            check([f"docs/{relative_path}.md", f"docs/{relative_path}/index.md"])
        elif relative_path.startswith("http://") or relative_path.startswith("https://"):
            # Don't check absolute paths to other locations
            return
        else:
            md_path = os.path.relpath(os.path.dirname(doc_path), repo_root)
            check(os.path.join(md_path, relative_path))
    else:
        if hasattr(parent, "children") and not isinstance(parent.children, str):
            for child in parent.children:
                check_local_links(child, doc_path, repo_root)
