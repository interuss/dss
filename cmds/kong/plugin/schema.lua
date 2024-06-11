local typedefs = require "kong.db.schema.typedefs"


return {
  name = "scope-acl",
  fields = {
    {
      service = typedefs.no_service,
    },
    {
      consumer = typedefs.no_consumer,
    },
    {
      protocols = typedefs.protocols_http,
    },
    {
      config = {
        type = "record",
        fields = {
            
        },
      },
    },
  },
}