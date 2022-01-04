import copy
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
    # print(cols)
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

  print('key_nodes: ', key_nodes)
  if len(key_nodes) >= 2:
    source = list(key_nodes)[0]
    destination = list(key_nodes)[1]
    # get path from source to destination
    all_paths = []
    get_path(parents, destination, source, [], all_paths)
    if not all_paths:
      # check if relation exists in reverse order.
      source, destination = destination, source
      get_path(parents, destination, source, [], all_paths)
      write_to_gv_file(key_nodes, all_paths)
    
  else:
    print('Source and destination modules not provided')


def write_to_gv_file(key_nodes, all_paths):
  with open('go_mod_graph2.gv', 'w') as f:
    f.write('digraph g {\n')
    f.write('  node [shape=box]\n')
    for kn in key_nodes:
        f.write('  "{}" [color=red]\n'.format(kn))
    pairs = []
    for path in all_paths:
        pairs.extend([(x,y) for x,y in zip(list(reversed(path))[:-1], list(reversed(path))[1:])])
    for pair in set(pairs):
        f.write('  "{}" -> "{}"\n'.format(pair[0], pair[1]))
    f.write('}\n')

def get_path(parents, dest, base, path, all_paths):
    if dest not in path:
        path.append(dest)
    if dest in parents:
        for parent in parents[dest]:
            curr_path = copy.copy(path)
            if parent not in curr_path:
                curr_path.append(parent)
            else:
                continue  # case of cyclic path.
            if parent == base:
                all_paths.append(curr_path)
            else:
                get_path(parents, parent, base, curr_path, all_paths)


if __name__ == '__main__':
  main()
