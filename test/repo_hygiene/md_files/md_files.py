import glob
import os

import marko

from .local_links import check_local_links


def check_md_file(md_file_path: str, repo_root: str) -> None:
    with open(md_file_path, "r") as f:
        doc = marko.parse(f.read())
    check_local_links(doc, md_file_path, repo_root)


def check_md_files(path: str, repo_root: str) -> None:
    for md_file in glob.glob(os.path.join(path, "*.md")):
        check_md_file(md_file, repo_root)
    for subfolder in (f.path for f in os.scandir(path) if f.is_dir()):
        check_md_files(subfolder, repo_root)
