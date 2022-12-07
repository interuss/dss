from implicitdict import ImplicitDict
from monitoring.uss_qualifier import fileio
from monitoring.uss_qualifier.resources.resource import Resource


class SourceDocumentSpecification(ImplicitDict):
    url: str
    """Url of the ED-269 document to verify"""


class SourceDocument(Resource[SourceDocumentSpecification]):
    specification: SourceDocumentSpecification

    raw_document: str
    """Content of the document"""

    def __init__(self, specification: SourceDocumentSpecification):
        self.specification = specification
        self.raw_document = fileio.load_content(specification.url)
