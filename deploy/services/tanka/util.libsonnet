{
  hyphenate(s): std.join('-', std.split(s, '_')),
  mapToList(o): [o[n] for n in std.objectFields(o)],
  objectItems(o): [[k, o[k]] for k in std.objectFields(o)],
  filter(o, fields): { [field]: o[field] for field in std.setInter(std.objectFields(o), fields) },
  exclude(o, fields): { [field]: o[field] for field in std.setDiff(std.objectFields(o), fields) },
  makeArgs(o): ['--%s=%s' % kv for kv in $.objectItems(o)],
  // Kubernetes container resources from the cpuRequest, memoryRequest, cpuLimit and memoryLimit fields of o, without the empty ones.
  resources(o):
    local requests = { [r]: o[r + 'Request'] for r in ['cpu', 'memory'] if o[r + 'Request'] != '' };
    local limits = { [r]: o[r + 'Limit'] for r in ['cpu', 'memory'] if o[r + 'Limit'] != '' };
    {
      [if requests != {} then 'requests']: requests,
      [if limits != {} then 'limits']: limits,
    },
}