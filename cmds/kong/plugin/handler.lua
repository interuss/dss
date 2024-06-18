
local ScopeAcl = {
    VERSION = "0.0.1",
    PRIORITY = 900
}

local groups = require "kong.plugins.acl.groups"

function ScopeAcl:access(config)
    kong.log("Started Authorization Service")

    local consumer_id = groups.get_current_consumer_id()
    local consumer_groups, err = groups.get_consumer_groups(consumer_id)
    if err then
        return error(err)
    end

    local required_scopes = kong.request.get_query()["scope"]

    for scope in string.gmatch(required_scopes, "%S+") do
        if consumer_groups[scope] == nil then
            return kong.response.error(403, string.format("You cannot request this scope: %s", scope))
        end
    end

end

return ScopeAcl