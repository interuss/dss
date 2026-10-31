"""Bundle file references into the local references understood by the generator."""

from pathlib import Path
from urllib.parse import unquote, urlsplit

import yaml


def load_spec(path: str | Path) -> dict:
    """Load an OpenAPI document and import the schemas it references.

    Each reference is resolved relative to its own document.  Schema references
    remain references (rather than being expanded) so shared and recursive types
    can be represented in Go.  Other referenced OpenAPI objects are expanded.
    """
    return _Resolver(Path(path).resolve()).bundle()


class _Resolver:
    def __init__(self, path: Path):
        self.path = path
        self.documents: dict[Path, dict] = {}
        self.schemas: dict[str, dict] = {}
        self.sources: dict[str, tuple[Path, tuple[str, ...]]] = {}

    def _document(self, path: Path) -> dict:
        if path not in self.documents:
            with path.open() as f:
                document = yaml.safe_load(f)
            if not isinstance(document, dict):
                raise ValueError(f"Expected an object in OpenAPI file {path}")
            self.documents[path] = document
        return self.documents[path]

    def _target(self, ref: str, source: Path) -> tuple[Path, tuple[str, ...], dict]:
        uri = urlsplit(ref)
        if uri.scheme or uri.netloc or uri.query:
            raise ValueError(
                f"Only local file $refs are supported; found {ref!r} in {source}"
            )
        path = (source.parent / unquote(uri.path)).resolve() if uri.path else source
        fragment = unquote(uri.fragment)
        if fragment and not fragment.startswith("/"):
            raise ValueError(f"Expected a JSON pointer in $ref {ref!r} in {source}")
        parts = (
            tuple(
                part.replace("~1", "/").replace("~0", "~")
                for part in fragment[1:].split("/")
            )
            if fragment
            else ()
        )
        try:
            target = self._document(path)
            for part in parts:
                target = target[int(part)] if isinstance(target, list) else target[part]
        except (OSError, KeyError, IndexError, TypeError, ValueError) as e:
            raise ValueError(f"Cannot resolve $ref {ref!r} in {source}: {e}") from e
        if not isinstance(target, dict):
            raise ValueError(f"Expected an object at $ref {ref!r} in {source}")
        return path, parts, target

    def _import_schema(self, path: Path, parts: tuple[str, ...], schema: dict) -> str:
        key = (path, parts)
        name = parts[-1] if parts else path.stem
        if name in self.sources and self.sources[name] != key:
            other_path, other_parts = self.sources[name]
            raise ValueError(
                f"Conflicting schema name {name!r}: "
                f"{path} ({'/'.join(parts) or 'root'}) and "
                f"{other_path} ({'/'.join(other_parts) or 'root'})"
            )
        self.sources[name] = key
        if name not in self.schemas:
            # Reserve the name before following references to support cycles.
            self.schemas[name] = {}
            self.schemas[name] = self._rewrite(schema, path, schema=True)
        return "#/components/schemas/" + name.replace("~", "~0").replace("/", "~1")

    def _rewrite(
        self,
        value,
        source: Path,
        schema: bool = False,
        expanding: tuple[tuple[Path, tuple[str, ...]], ...] = (),
    ):
        if isinstance(value, list):
            return [self._rewrite(item, source, schema, expanding) for item in value]
        if not isinstance(value, dict):
            return value
        if "$ref" in value:
            path, parts, target = self._target(value["$ref"], source)
            siblings = self._rewrite(
                {k: v for k, v in value.items() if k != "$ref"},
                source,
                schema,
                expanding,
            )
            if schema or parts[:2] == ("components", "schemas"):
                return {"$ref": self._import_schema(path, parts, target), **siblings}
            key = (path, parts)
            if key in expanding:
                raise ValueError(
                    f"Cyclic non-schema $ref {value['$ref']!r} in {source}"
                )
            return {
                **self._rewrite(target, path, expanding=(*expanding, key)),
                **siblings,
            }
        result = {}
        for key, item in value.items():
            if (
                key == "example"
                or (isinstance(key, str) and key.startswith("x-"))
                or (schema and key in {"default", "enum", "const", "examples"})
            ):
                # These values are instance data and may contain literal $refs.
                result[key] = item
            elif schema and key in {
                "properties",
                "patternProperties",
                "$defs",
                "definitions",
            }:
                result[key] = {
                    name: self._rewrite(definition, source, True, expanding)
                    for name, definition in item.items()
                }
            else:
                result[key] = self._rewrite(
                    item, source, schema or key == "schema", expanding
                )
        return result

    def bundle(self) -> dict:
        document = self._document(self.path)
        components = document.get("components", {})
        # Reserve all root names before importing any external schemas.
        for name in components.get("schemas", {}):
            self.sources[name] = (self.path, ("components", "schemas", name))
        for name, schema in components.get("schemas", {}).items():
            self._import_schema(self.path, ("components", "schemas", name), schema)
        result = self._rewrite(
            {
                **document,
                "components": {k: v for k, v in components.items() if k != "schemas"},
            },
            self.path,
        )
        # Keep the root declaration order so existing generated files stay stable.
        schemas = {name: self.schemas[name] for name in components.get("schemas", {})}
        schemas.update(self.schemas)
        result["components"]["schemas"] = schemas
        return result
