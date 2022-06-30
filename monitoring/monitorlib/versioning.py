import os
import subprocess


def get_code_version() -> str:
  env_version = os.environ.get('MONITORING_VERSION', '')
  if env_version:
    return env_version
  env_version = os.environ.get('CODE_VERSION', '')
  if env_version:
    return env_version

  process = subprocess.Popen(['git', 'rev-parse', '--short', 'HEAD'],
                             stdout=subprocess.PIPE,
                             universal_newlines=True)
  commit, _ = process.communicate()
  if process.returncode != 0:
    return 'unknown'
  commit = commit.strip()
  if 'not a git repository' in commit:
    return 'unknown'

  process = subprocess.Popen(['git', 'status', '-s'],
                             stdout=subprocess.PIPE,
                             universal_newlines=True)
  status, _ = process.communicate()
  if process.returncode != 0:
    return commit + '-unknown'
  elif status:
    return commit + '-dirty'
  return commit
