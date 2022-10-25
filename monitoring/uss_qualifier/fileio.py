import json
import os
from typing import Tuple, Optional, Dict

import requests
import yaml

FILE_PREFIX = "file://"
HTTP_PREFIX = "http://"
HTTPS_PREFIX = "https://"
RECOGNIZED_EXTENSIONS = ".json", ".yaml", ".kml"


FileReference = str
"""Location of a file containing content.

May be:
  * file://<PATH>
  * http(s)://<PATH>
  * Python-package style name relative to the uss_qualifier package (without extension; extension will be inferred by what file is present)

Allowed extensions:
  * .json (dict, content)
  * .yaml (dict, content)
  * .kml (content)
"""


def resolve_filename(data_file: FileReference) -> str:
    if data_file.startswith(FILE_PREFIX):
        # file:// explicit local file reference
        return os.path.abspath(data_file[len(FILE_PREFIX) :])
    elif data_file.startswith(HTTP_PREFIX) or data_file.startswith(HTTPS_PREFIX):
        # http(s):// web file reference
        return data_file
    else:
        # Package-based name (without extension)
        path_parts = [os.path.dirname(__file__)]
        path_parts += data_file.split(".")
        file_name = None

        for ext in RECOGNIZED_EXTENSIONS:
            ext_file = os.path.join(*path_parts) + ext
            if os.path.exists(ext_file):
                return os.path.abspath(ext_file)

        if file_name is None:
            raise NotImplementedError(
                f"Cannot load find a suitable file to load for {data_file}"
            )


def _load_content_from_file_name(file_name: str) -> str:
    if file_name.startswith(HTTP_PREFIX) or file_name.startswith(HTTPS_PREFIX):
        # http(s):// web file reference
        resp = requests.get(file_name)
        resp.raise_for_status()
        file_content = resp.content.decode("utf-8")
    else:
        with open(file_name, "r") as f:
            file_content = f.read()

    return file_content


def load_content(data_file: FileReference) -> str:
    return _load_content_from_file_name(resolve_filename(data_file))


def _split_anchor(file_name: str) -> Tuple[str, Optional[str]]:
    if "#" in file_name:
        anchor_location = file_name.index("#")
        base_file_name = file_name[0:anchor_location]
        anchor = file_name[anchor_location + 1 :]
    else:
        base_file_name = file_name
        anchor = None
    return base_file_name, anchor


def load_dict(data_file: FileReference) -> dict:
    """Loads a dict from the specified file reference.

    If the data_file has a #<COMPONENT_PATH> suffix, the component at that path
    will be selected.  For example, #/foo/bar will select content["foo"]["bar"].

    If any key at any level of the loaded content is "$ref", then the value is
    expected to be a string that refers to a file (plus, optionally, a component
    path following #), or a blank string followed by a # component path which
    refers to a component within the current file.  All of the keys from the
    $ref will be added to the parent object of the $ref, and the $ref will be
    removed.  This $ref convention is generally compatible with OpenAPI, except
    that other keys may co-exist with $ref.  Multiple $refs may be used when
    enclosed in an allOf key (with an array as a value), again similar to
    OpenAPI.
    """
    base_file_name, anchor = _split_anchor(data_file)
    base_file_name = resolve_filename(base_file_name)
    file_name = base_file_name + (f"#{anchor}" if anchor is not None else "")
    return _load_dict_from_file_name(file_name, file_name)


def _load_dict_from_file_name(
    file_name: str, context_file_name: str, cache: Optional[Dict[str, dict]] = None
) -> dict:
    if cache is None:
        cache = {}

    base_file_name, anchor = _split_anchor(file_name)

    if base_file_name.startswith(FILE_PREFIX):
        base_file_name = base_file_name[len(FILE_PREFIX) :]
    if (
        not base_file_name.startswith(HTTP_PREFIX)
        and not base_file_name.startswith(HTTPS_PREFIX)
        and not base_file_name.startswith("/")
    ):
        # This is a relative file path; it should be relative to the context
        root_path = os.path.dirname(context_file_name)
        base_file_name = os.path.join(root_path, base_file_name)

    if base_file_name in cache:
        dict_content = cache[base_file_name]
    else:
        file_content = _load_content_from_file_name(base_file_name)

        if base_file_name.lower().endswith(".json"):
            dict_content = json.loads(file_content)
        elif base_file_name.lower().endswith(".yaml"):
            dict_content = yaml.safe_load(file_content)
        else:
            raise NotImplementedError(
                f'Unable to parse data for "{base_file_name}" because its extension-based data format is not supported'
            )

        _replace_refs(dict_content, dict_content, base_file_name, cache)
        _drop_allof(dict_content)
        cache[base_file_name] = dict_content

    if anchor is not None:
        return _select_path(dict_content, anchor)
    else:
        return dict_content


def _replace_refs(
    content: dict, context_content: dict, context_file_name: str, cache: Dict[str, dict]
) -> None:
    if "$ref" in content:
        ref_path = content.pop("$ref")
        if not isinstance(ref_path, str):
            raise ValueError(
                f"$ref link must be a string; found instead: {str(ref_path)}"
            )
        if ref_path.startswith("#"):
            ref_content = _select_path(context_content, ref_path[1:])
        else:
            ref_content = _load_dict_from_file_name(ref_path, context_file_name, cache)
        for k, v in ref_content.items():
            content[k] = v
        _replace_refs(content, context_content, context_file_name, cache)
    else:
        for v in content.values():
            if isinstance(v, dict):
                _replace_refs(v, context_content, context_file_name, cache)
            try:
                iterable = iter(v)
            except TypeError:
                iterable = None
            if iterable:
                for item in v:
                    if isinstance(item, dict):
                        _replace_refs(item, context_content, context_file_name, cache)


def _select_path(content: dict, path: str) -> dict:
    if not path.startswith("/"):
        raise ValueError(
            f'Relative path to dict component must start with /; found instead: "{path}"'
        )
    path = path[1:]
    if "/" not in path:
        if not path in content:
            raise KeyError(
                f'Could not find key "{path}" in dict; found keys: {", ".join(content)}'
            )
        return content[path]
    else:
        separator_location = path.index("/")
        component = path[0:separator_location]
        subpath = path[separator_location:]
        if component not in content:
            raise KeyError(
                f'Could not find key "{path}" in dict content: {str(content)}'
            )
        return _select_path(content[component], subpath)


def _drop_allof(content: dict) -> None:
    if "allOf" in content:
        all_of = content.pop("allOf")
        for schema in all_of:
            _drop_allof(schema)
            for k, v in schema.items():
                content[k] = v
    for v in content.values():
        if isinstance(v, dict):
            _drop_allof(v)
        try:
            iterable = iter(v)
        except TypeError:
            iterable = None
        if iterable:
            for item in v:
                if isinstance(item, dict):
                    _drop_allof(item)
