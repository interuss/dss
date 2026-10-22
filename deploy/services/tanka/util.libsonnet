{
  hyphenate(s): std.join('-', std.split(s, '_')),
  mapToList(o): [o[n] for n in std.objectFields(o)],
  objectItems(o): [[k, o[k]] for k in std.objectFields(o)],
  filter(o, fields): { [field]: o[field] for field in std.setInter(std.objectFields(o), fields) },
  exclude(o, fields): { [field]: o[field] for field in std.setDiff(std.objectFields(o), fields) },
  makeArgs(o): ['--%s=%s' % kv for kv in $.objectItems(o)],

  // deterministically derive an int from a string
  stringToInt(str, mod):
    local hash = std.md5(str);
    local int = std.parseHex(hash[0:8]); // parse first 8 hex characters (32 bits) to integer
    std.mod(int, mod)
}
