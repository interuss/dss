# Rewrites links to the master branch of the repository to DOCS_GIT_REF (the release tag) when set,
# so that the documentation of a release links to the files of that release.

import os
import re

_MASTER_LINK = re.compile(r"(https://github\.com/interuss/dss/(?:blob|tree))/master/")


def on_page_markdown(markdown, **kwargs):
    ref = os.environ.get("DOCS_GIT_REF")
    if not ref:
        return markdown
    return _MASTER_LINK.sub(lambda m: f"{m.group(1)}/{ref}/", markdown)
