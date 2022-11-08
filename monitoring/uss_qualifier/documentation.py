import marko.element


def text_of(parent: marko.element.Element) -> str:
    if not hasattr(parent, "children"):
        return ""
    if isinstance(parent.children, str):
        return parent.children
    return "".join(text_of(c) for c in parent.children)
