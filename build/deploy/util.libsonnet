{
  hyphenate(s): std.join('-', std.split(s, '_')),
  mapToList(o): [o[n] for n in std.objectFields(o)],
  objectItems(o): [[k, o[k]] for k in std.objectFields(o)],
  filter(o, fields): { [field]: o[field] for field in std.setInter(std.objectFields(o), fields) },
  exclude(o, fields): { [field]: o[field] for field in std.setDiff(std.objectFields(o), fields) },
}