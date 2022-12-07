from typing import Optional, List, Set, Tuple, Dict

import graphviz

from implicitdict import ImplicitDict

from monitoring.uss_qualifier.reports.report import (
    ActionGeneratorReport,
    TestRunReport,
    TestSuiteReport,
    TestScenarioReport,
)
from monitoring.uss_qualifier.resources.definitions import (
    ResourceID,
    ResourceCollection,
)
from monitoring.uss_qualifier.scenarios.definitions import TestScenarioDeclaration
from monitoring.uss_qualifier.suites.definitions import (
    ActionType,
    TestSuiteDeclaration,
    TestSuiteDefinition,
    ActionGeneratorDefinition,
)

NodeName = str


class Node(ImplicitDict):
    """Represents a node to be used in a GraphViz graph."""

    name: NodeName
    label: Optional[str] = None
    children: List[NodeName]
    attributes: Dict[str, str]


class NodeNamer(object):
    """Creates and tracks unique, valid names within a GraphViz graph."""

    names: Set[NodeName] = set()

    def make_name(self, desired_name: NodeName) -> NodeName:
        acceptable_name = desired_name.replace(" ", "").replace(".", "_")
        actual_name = acceptable_name
        i = 2
        while actual_name in self.names:
            actual_name = f"{acceptable_name}_{i}"
            i += 1
        self.names.add(actual_name)
        return actual_name

    def use_name(self, name: NodeName) -> None:
        self.names.add(name)


def _make_test_scenario_nodes(
    declaration: Optional[TestScenarioDeclaration],
    report: TestScenarioReport,
    nodes_by_id: Dict[ResourceID, Node],
    namer: NodeNamer,
    include_notes: bool = False,
) -> List[Node]:
    nodes: List[Node] = []

    # Make the scenario node
    scenario_type = report.scenario_type
    if scenario_type.startswith("scenarios."):
        scenario_type = scenario_type[len("scenarios.") :]
    label_elements = [report.name, scenario_type]
    if include_notes and "notes" in report:
        label_elements += [f"{k}={v.message}" for k, v in report.notes.items()]
    scenario_node = Node(
        name=namer.make_name(report.scenario_type),
        label="\n".join(label_elements),
        children=[],
        attributes={
            "shape": "component",
            "fillcolor": "lightgreen" if report.successful else "lightpink",
            "style": "filled",
        },
    )

    # Mark the scenario node a child of the appropriate resources
    if declaration is not None:
        for local_id, node in nodes_by_id.items():
            node.children.append(scenario_node.name)

    # Add failed checks and error below
    parent_node = scenario_node
    for failed_check in report.get_all_failed_checks():
        check_node = Node(
            name=namer.make_name(report.scenario_type + "FailedCheck"),
            label=failed_check.summary,
            children=[],
            attributes={
                "shape": "octagon",
                "fillcolor": "lightpink",
                "style": "filled",
            },
        )
        nodes.append(check_node)
        parent_node.children.append(check_node.name)
        parent_node = check_node
    if "execution_error" in report:
        error_node = Node(
            name=namer.make_name(report.scenario_type + "ExecutionError"),
            label=report.execution_error.message,
            children=[],
            attributes={
                "shape": "octagon",
                "fillcolor": "lightpink",
                "style": "filled",
                "color": "red",
            },
        )
        nodes.append(error_node)
        parent_node.children.append(error_node.name)

    # Add the scenario node last
    nodes.append(scenario_node)

    return nodes


def _make_test_suite_nodes(
    declaration: TestSuiteDeclaration,
    report: TestSuiteReport,
    nodes_by_id: Dict[ResourceID, Node],
    namer: NodeNamer,
) -> List[Node]:
    # Make child nodes for each action in the suite
    new_nodes = None
    nodes: List[Node] = []
    children: List[NodeName] = []
    definition = TestSuiteDefinition.load(declaration.suite_type)
    for action_report, action in zip(report.actions, definition.actions):
        action_nodes = _translate_ids(nodes_by_id, action.get_resource_links())
        if "test_suite" in action_report:
            new_nodes = _make_test_suite_nodes(
                action.test_suite,
                action_report.test_suite,
                action_nodes,
                namer,
            )
            children.append(new_nodes[-1].name)
        elif "test_scenario" in action_report:
            new_nodes = _make_test_scenario_nodes(
                action.test_scenario, action_report.test_scenario, action_nodes, namer
            )
            nodes.extend(new_nodes)
            children.append(new_nodes[-1].name)
        elif "action_generator" in action_report:
            new_nodes = _make_action_generator_nodes(
                action.action_generator,
                action_report.action_generator,
                action_nodes,
                namer,
            )
            children.append(new_nodes[-1].name)
        else:
            ActionType.raise_invalid_action_declaration()
        nodes.extend(new_nodes)

    # Make the suite node itself
    suite_type = report.suite_type
    if suite_type.startswith("suites."):
        suite_type = suite_type[len("suites.") :]
    suite_node = Node(
        name=namer.make_name(report.suite_type),
        label=report.name + "\n" + suite_type,
        children=children,
        attributes={"shape": "folder"},
    )
    nodes.append(suite_node)
    return nodes


def _make_action_generator_nodes(
    definition: ActionGeneratorDefinition,
    report: ActionGeneratorReport,
    nodes_by_id: Dict[ResourceID, Node],
    namer: NodeNamer,
) -> List[Node]:
    # Make the action generator node
    nodes: List[Node] = []
    children: List[NodeName] = []
    for action in report.actions:
        if "test_suite" in action:
            raise NotImplementedError()
        elif "test_scenario" in action:
            new_nodes = _make_test_scenario_nodes(
                None, action.test_scenario, nodes_by_id, namer, True
            )
            nodes.extend(new_nodes)
            children.append(new_nodes[-1].name)
        elif "action_generator" in action:
            raise NotImplementedError()
        else:
            raise NotImplementedError()
        nodes.extend(new_nodes)
    generator_type = report.generator_type
    if generator_type.startswith("action_generators."):
        generator_type = generator_type[len("action_generators.") :]
    generator_node = Node(
        name=namer.make_name(report.generator_type),
        label="Action generator\n" + generator_type,
        children=children,
        attributes={"shape": "box3d"},
    )
    nodes.append(generator_node)

    # Point all resources used by children of the action generator to the action generator
    for local_id, node in _translate_ids(nodes_by_id, definition.resources).items():
        node.children.append(generator_node.name)

    return nodes


def _make_resource_nodes(
    resources: ResourceCollection, namer: NodeNamer
) -> Tuple[List[Node], Dict[ResourceID, Node]]:
    nodes: List[Node] = []
    nodes_by_id: Dict[ResourceID, Node] = {}
    added = 1
    while added > 0:
        added = 0
        for resource_id, declaration in resources.resource_declarations.items():
            # Skip resources that already have nodes
            if resource_id in namer.names:
                continue

            # Identify prerequisite resources not yet created
            prereqs = []
            for child_param, parent_id in declaration.dependencies.items():
                if parent_id not in nodes_by_id:
                    prereqs.append(parent_id)

            # Create this resource node if there are no uncreated prerequisites
            if not prereqs:
                resource_type = declaration.resource_type
                if resource_type.startswith("resources."):
                    resource_type = resource_type[len("resources.") :]
                resource_node = Node(
                    name=namer.make_name(resource_id),
                    label=f"{resource_id}\n{resource_type}",
                    children=[],
                    attributes={
                        "shape": "note",
                        "fillcolor": "lightskyblue1",
                        "style": "filled",
                    },
                )
                namer.use_name(resource_id)
                nodes.append(resource_node)
                nodes_by_id[resource_id] = resource_node
                for child_param, parent_id in declaration.dependencies.items():
                    nodes_by_id[parent_id].children.append(resource_node.name)
                added += 1
    return nodes, nodes_by_id


def _translate_ids(
    nodes_by_id: Dict[ResourceID, Node], local_resources: Dict[ResourceID, ResourceID]
) -> Dict[ResourceID, Node]:
    local_id_by_parent_id = {v: k for k, v in local_resources.items()}
    return {
        local_id_by_parent_id[parent_id]: node
        for parent_id, node in nodes_by_id.items()
        if parent_id in local_id_by_parent_id
    }


def make_graph(report: TestRunReport) -> graphviz.Digraph:
    namer = NodeNamer()

    # Make nodes for resources
    nodes, nodes_by_id = _make_resource_nodes(report.configuration.resources, namer)

    if report.configuration.action.get_action_type() != ActionType.TestSuite:
        raise NotImplementedError()
    test_suite = report.configuration.action.test_suite
    test_suite_report = report.report.test_suite

    # Translate resource names into the action frame
    suite_nodes_by_id = _translate_ids(nodes_by_id, test_suite.resources)

    # Make nodes for the suite
    nodes.extend(
        _make_test_suite_nodes(test_suite, test_suite_report, suite_nodes_by_id, namer)
    )

    # Translate nodes into GraphViz
    dot = graphviz.Digraph(node_attr={"shape": "box"})
    for node in nodes:
        dot.node(name=node.name, label=node.label, **node.attributes)
        for child in node.children:
            dot.edge(node.name, child)

    return dot
