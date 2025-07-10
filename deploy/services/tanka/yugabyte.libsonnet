# TODO: This file is not implemented yet
local base = import 'base.libsonnet';
local util = import 'util.libsonnet';
local volumes = import 'volumes.libsonnet';

{
  all(metadata): if metadata.datastore == 'yugabyte' then {
  } else {}
}
