{
  mapToList(o): [o[n] for n in std.objectFields(o)],
  objectItems(o): [[k, o[k]] for k in std.objectFields(o)],
  exclude(o, fields): { [field]: o[field] for field in std.setDiff(std.objectFields(o), fields) },
  makeArgs(o): ['--%s=%s' % kv for kv in $.objectItems(o)],
  makeEnv(o): [{ name: kv[0], value: std.toString(kv[1]) } for kv in $.objectItems(o)],
  // Kubernetes container resources from the cpuRequest, memoryRequest, cpuLimit and memoryLimit fields of o, without the empty ones.
  resources(o):
    local requests = { [r]: o[r + 'Request'] for r in ['cpu', 'memory'] if o[r + 'Request'] != '' };
    local limits = { [r]: o[r + 'Limit'] for r in ['cpu', 'memory'] if o[r + 'Limit'] != '' };
    {
      [if requests != {} then 'requests']: requests,
      [if limits != {} then 'limits']: limits,
    },
  // deterministically derive an int from a string
  stringToInt(str, mod):
    local hash = std.md5(str);
    local int = std.parseHex(hash[0:8]); // parse first 8 hex characters (32 bits) to integer
    std.mod(int, mod)
}
