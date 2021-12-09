import subprocess
import sys
from typing import Dict, Set, Tuple

"""
This script produces a graph indicating what dependencies lead to specified
modules being included in the Go distribution.

Usage:
  python why_go_dep.py github.com/example/module-1 github.com/example/module-2 [...]
The arguments are a list of modules of interest.  These modules and their
ancestor dependencies will be included in the output graph.

Output:
go_mod_graph.gv: Graphviz descriptor of the module dependencies
go_mod_graph.png: Produced if `dot` is present on the system, a Graphviz
  rendering of go_mod_graph.gv.
"""

def main():
  keywords = sys.argv

  lines = subprocess.check_output(['go', 'mod', 'graph']).decode('utf8').split('\n')

  # `connections` contains a set of tuples describing a parent (first element)
  # connection to a child (second element)
  connections: Set[Tuple[str, str]] = set()

  # `parents` maps the name of a child to a set of the names of its parents
  parents: Dict[str, Set[str]] = {}

  # `children` maps the name of a parent to a set of the names of its children
  children: Dict[str, Set[str]] = {}

  for line in lines:
    cols = line.split(' ')
    if len(cols) != 2:
      continue
    parent, child = cols
    parent = parent.split('@')[0]
    child = child.split('@')[0]

    connections.add((parent, child))

    deps = children.get(parent, set())
    deps.add(child)
    children[parent] = deps

    parent_set = parents.get(child, set())
    parent_set.add(parent)
    parents[child] = parent_set

  # Identify all full node names that match a specified module of interest
  key_nodes: Set[str] = set()
  for child in children.keys():
    for kw in keywords:
      if kw in child:
        key_nodes.add(child)
  for parent in parents.keys():
    for kw in keywords:
      if kw in parent:
        key_nodes.add(parent)

  # Nodes are relevant if one of their descendents is a module of interest
  relevant: Set[str] = {n for n in key_nodes}
  newly_relevant: Set[str] = {n for n in relevant}
  while len(newly_relevant) > 0:
    next_relevant: Set[str] = set()
    for n in newly_relevant:
      for p in children.get(n, set()):
        if p not in relevant:
          relevant.add(p)
          next_relevant.add(p)
    newly_relevant = next_relevant

  with open('go_mod_graph.gv', 'w') as f:
    f.write('digraph g {\n')
    f.write('  node [shape=box]\n')
    for kn in key_nodes:
      f.write('  "{}" [color=red]\n'.format(kn))
    for connection in connections:
      if connection[0] in relevant and connection[1] in relevant:
        f.write('  "{}" -> "{}"\n'.format(connection[0], connection[1]))
    f.write('}\n')

  # TODO: check if `dot` is present in the system.
  # img_gen = subprocess.check_output(['dot', '-Tpng', '-ogo_mod_graph.png', 'go_mod_graph.gv']).decode('utf8')
  # print(img_gen)


if __name__ == '__main__':
  main()
