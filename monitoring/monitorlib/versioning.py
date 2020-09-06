import os
import subprocess


def get_code_version() -> str:
  if os.path.exists('VERSION'):
    with open('VERSION', 'r') as f:
      return f.read()

  process = subprocess.Popen(['git', 'rev-parse', '--short', 'HEAD'],
                             stdout=subprocess.PIPE,
                             universal_newlines=True)
  commit, _ = process.communicate()
  if process.returncode != 0:
    return 'unknown'
  commit = commit.strip()

  process = subprocess.Popen(['git', 'status', '-s'],
                             stdout=subprocess.PIPE,
                             universal_newlines=True)
  status, _ = process.communicate()
  if process.returncode != 0:
    return commit + '-unknown'
  elif status:
    return commit + '-dirty'
  return commit
