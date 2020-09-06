from typing import Dict, Optional

from termcolor import colored
import yaml


def isa_diff_text(a: Optional[Dict], b: Optional[Dict]) -> str:
  if a and not b:
    return colored(yaml.dump(a), 'red')
  elif b and not a:
    return colored(yaml.dump(b), 'green')

  lines = []
  for k1, v1 in b.items():
    if k1 not in a:
      lines.append(colored('{}: {}'.format(k1, v1), 'green'))
    elif v1 != a[k1]:
      lines.append(colored('{}: {}'.format(k1, v1), 'yellow'))
  for k0, v0 in a.items():
    if k0 not in b:
      lines.append(colored('{}: {}'.format(k0, v0), 'red'))
  return '\n'.join(lines)
