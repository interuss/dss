import re

import apis
import data_types
import operations


def comment(lines: list[str]) -> list[str]:
    """Prepend comment characters to each line.

    :param lines: Lines of text to be commented
    :return: Same lines of text provided after each line is commented
    """
    return ["// " + line for line in lines]


def indent(lines: list[str], level: int) -> list[str]:
    """Indent each line.

    :param lines: Lines of text to be indented
    :param level: Level of indent (each indent level is two spaces)
    :return: Same lines of text provided after each line is indented
    """
    if level == 0:
        return lines
    else:
        return ["  " * level + line if line else "" for line in lines]


def imports(import_list: list[str]) -> str:
    return "\n".join(
        indent([f'"{i}"' if '"' not in i else i for i in import_list], 2)
    )


def template_content(template_name: str, template_vars: dict[str, str]) -> str:
    """Fill in a template with provided values and return the entire content.

    :param template_name: Name of template file in `templates` folder (e.g., 'common' reads from `templates/common.go.template`)
    :param template_vars: Mapping of key (sentinel in template) to value (what to replace the sentinel with)
    :return: Template content with filled values
    """
    with open(f"templates/{template_name}.go.template") as f:
        content = f.read()
    for k, v in template_vars.items():
        content = content.replace(k, v)
    return content


def data_type(d_type: data_types.DataType) -> list[str]:
    """Generate Go code defining the provided data type.

    :param d_type: Parsed API data type to render into Go code
    :return: Lines of Go code defining the provided data type
    """
    lines = comment(d_type.description.split("\n")) if d_type.description else []

    if data_types.is_primitive_go_type(d_type.go_type):
        lines.append(f"type {d_type.name} {d_type.go_type}")
    elif d_type.go_type == "struct":
        lines.append("type %s struct {" % d_type.name)
        for field in d_type.fields:
            lines.extend(indent(_object_field(field), 1))
            lines.append("")
        if d_type.fields:
            lines.pop()
            lines.append("}")
        else:
            lines[-1] += "}"
    else:
        lines.append(f"type {d_type.name} {d_type.go_type}")

    if d_type.enum_values:
        lines.append("const (")
        lines.extend(
            indent(
                [
                    f'{d_type.name}_{v} {d_type.name} = "{v}"'
                    for v in d_type.enum_values
                ],
                1,
            )
        )
        lines.append(")")

    return lines


def _object_field(field: data_types.ObjectField) -> list[str]:
    """Generate an unindented definition of the provided field in Go code.

    :param field: Data type field to render into Go code
    :return: Lines of Go code defining the provided field
    """
    lines = comment(field.description.split("\n")) if field.description else []
    lines.append(
        '{} {}{} `json:"{}{}"`'.format(
            field.go_name,
            "*" if not field.required else "",
            field.go_type,
            field.api_name,
            ",omitempty" if not field.required else "",
        )
    )
    return lines


def implementation_interface(
    api: apis.API, api_package: str, ensure_500: bool
) -> list[str]:
    """Generate Go code defining the interface an API implementation must implement.

    :param api: API to be rendered into an interface
    :param api_package: Name of root/common API package
    :param ensure_500: If True, add a 500 response to all operations that don't already define a 500 response
    :return: Lines of Go code defining the interface
    """
    lines: list[str] = []

    # Provide security constants
    lines.append("var (")

    for _, scope in api.security_scopes():
        lines.append(
            '%s = api.RequiredScope("%s")' % (scope.go_constant_name, scope.name)
        )

    var_body: list[str] = []
    for operation in api.operations:
        var_body.append(
            "%sSecurity = []%s.AuthorizationOption{"
            % (operation.interface_name, api_package)
        )

        init_body: list[str] = []
        for auth_option in operation.security.options:
            init_body.append("{")
            for scheme, scopes in auth_option.option.items():
                init_body.extend(
                    indent(
                        [
                            '"%s": {%s},'
                            % (
                                scheme,
                                ", ".join(scope.go_constant_name for scope in scopes),
                            )
                        ],
                        1,
                    )
                )
            init_body.append("},")
        var_body.extend(indent(init_body, 1))

        var_body.append("}")
    lines.extend(indent(var_body, 1))

    lines.append(")")

    # Declare request & response types for all operations
    for operation in api.operations:
        lines.append("")

        # Declare request type for operation
        lines.append(f"type {operation.request_type_name} struct {{")

        body: list[str] = []
        for p in operation.path_parameters + operation.query_parameters:
            if p.description:
                body.extend(comment(p.description.split("\n")))
            body.append(
                "{} {}{}".format(
                    p.go_field_name,
                    "*" if p in operation.query_parameters else "",
                    p.go_type,
                )
            )
            body.append("")
        if operation.json_request_body_type:
            body.extend(
                comment(
                    [
                        "The data contained in the body of this request, if it parsed correctly"
                    ]
                )
            )
            body.append(f"Body *{operation.json_request_body_type}")
            body.append("")
            body.extend(
                comment(
                    [
                        "The error encountered when attempting to parse the body of this request"
                    ]
                )
            )
            body.append("BodyParseError error")
            body.append("")
        body.extend(comment(["The result of attempting to authorize this request"]))
        body.append(f"Auth {api_package}.AuthorizationResult")
        lines.extend(indent(body, 1))

        lines.append("}")
        lines.append("")

        lines.append(
            f'const {operation.interface_name}OperationID = "{api.package}.{operation.interface_name}"'
        )
        lines.append("")
        lines.append(
            f"func (request *{operation.request_type_name}) OperationID() string {{ return {operation.interface_name}OperationID }}"
        )

        # Declare response type for operation
        lines.append(f"type {operation.response_type_name} struct {{")

        body: list[str] = []
        responses = [r for r in operation.responses]
        if ensure_500 and 500 not in {r.code for r in responses}:
            responses.append(
                operations.Response(
                    code=500,
                    description="Auto-generated internal server error response",
                    json_body_type=f"{api_package}.InternalServerErrorBody",
                )
            )
        for response in responses:
            if response.description:
                body.extend(comment(response.description.split("\n")))
            body_type = (
                response.json_body_type
                if response.json_body_type
                else f"{api_package}.EmptyResponseBody"
            )
            body.extend([f"{response.response_set_field} *{body_type}"])
            body.append("")
        body.pop()
        lines.extend(indent(body, 1))

        lines.append("}")

    lines.append("")
    lines.append("type Implementation interface {")

    body: list[str] = []
    for operation in api.operations:
        comments: list[str] = []
        if operation.summary and operation.summary != operation.description:
            comments.extend(operation.summary.split("\n"))
        if (
            operation.summary
            and operation.description
            and operation.summary != operation.description
        ):
            comments.append("---")
        if operation.description:
            comments.extend(operation.description.split("\n"))
        body.extend(comment(comments))

        body.append(
            f"{operation.interface_name}(ctx context.Context, req *{operation.request_type_name}) {operation.response_type_name}"
        )
        body.append("")
    if body:
        body.pop()
    lines.extend(indent(body, 1))

    lines.append("}")
    return lines


def routes(
    api: apis.API, api_package: str, ensure_500: bool
) -> tuple[list[str], set[str]]:
    """Generate handler Go code for each operation, routed appropriately.

    :param api: API to have its operation routes rendered
    :param api_package: Name of root/common API package
    :param ensure_500: If True, add a 500 response to all operations that don't already define a 500 response
    :return:
        * Lines of Go code defining the handler functions and router creation function
        * Go packages that need to be imported
    """
    lines: list[str] = []
    imports: set[str] = set()
    need_auth_error = False

    # Define a top-level routed HTTP handler function for each operation
    for operation in api.operations:
        imports.add("regexp")
        imports.add("net/http")
        lines.append(
            f"func (s *APIRouter) {operation.interface_name}(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {{"
        )

        body: list[str] = []

        # Create object to hold the processed input and output to the operation
        body.append(f"var req {operation.request_type_name}")
        body.append(f"var response {operation.response_type_name}")
        body.append("")

        # Parse any path parameters
        if operation.path_parameters:
            body.extend(comment(["Parse path parameters"]))
            body.append("pathMatch := exp.FindStringSubmatch(r.URL.Path)")
            for i, p in enumerate(operation.path_parameters):
                if p.go_type == "string":
                    body.append(f"req.{p.go_field_name} = pathMatch[{i + 1}]")
                else:
                    body.append(
                        f"req.{p.go_field_name} = {p.go_type}(pathMatch[{i + 1}])"
                    )
            body.append("")

        # Capture/parse any query parameters
        if operation.query_parameters:
            body.extend(comment(["Copy query parameters"]))
            body.append("query := r.URL.Query()")
            for q in operation.query_parameters:
                body.append('if query.Has("%s") {' % q.name)
                if_body: list[str] = []
                if q.go_type == "string":
                    if_body.append(f'v := query.Get("{q.name}")')
                    if_body.append(f"req.{q.go_field_name} = &v")
                else:
                    primitive_type = api.primitive_go_type_for(q.go_type)
                    if primitive_type == "string":
                        if_body.append(
                            f'v := {q.go_type}(query.Get("{q.name}"))'
                        )
                        if_body.append(f"req.{q.go_field_name} = &v")
                    elif primitive_type.startswith("int") or primitive_type.startswith(
                        "float"
                    ):
                        imports.add("strconv")
                        if primitive_type.startswith("int"):
                            parse_func = "ParseInt"
                            parse_params = "10, {}".format(primitive_type[len("int") :])
                            no_conversion_type = "int64"
                        elif primitive_type.startswith("float"):
                            parse_func = "ParseFloat"
                            parse_params = primitive_type[len("float") :]
                            no_conversion_type = "float64"
                        else:
                            raise RuntimeError()
                        if_body.append(
                            f'i, err := strconv.{parse_func}(query.Get("{q.name}"), {parse_params})'
                        )
                        if_body.append("if err == nil {")
                        if q.go_type == no_conversion_type:
                            if_body.append(f"req.{q.go_field_name} = &i")
                        else:
                            if_body.extend(
                                indent(
                                    [
                                        f"v := {q.go_type}(i)",
                                        f"req.{q.go_field_name} = &v",
                                    ],
                                    1,
                                )
                            )
                        if_body.append("}")
                    else:
                        raise NotImplementedError()
                body.extend(indent(if_body, 1))
                body.append("}")
            body.append("")

        # Attempt to parse the request body JSON, if defined
        if operation.json_request_body_type:
            imports.add("encoding/json")
            body.extend(comment(["Parse request body"]))
            body.append(f"req.Body = new({operation.json_request_body_type})")
            body.append("defer r.Body.Close()")
            body.append("req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)")
            body.append("")

        # Actually invoke the API Implementation with the processed request to obtain the response
        imports.add("context")

        body.append("")

        call_block = []

        call_block.extend(comment(["Call implementation"]))
        call_block.append("ctx, cancel := context.WithCancel(r.Context())")
        call_block.append("defer cancel()")
        call_block.append(
            f"response = s.Implementation.{operation.interface_name}(ctx, &req)"
        )

        if operation.security.options:
            need_auth_error = True
            # Authorize & verify the call
            body.extend(comment(["Authorize request"]))
            body.append(
                f"req.Auth = s.Authorizer.Authorize(w, r, {operation.interface_name}Security)"
            )

            body.extend(comment(["Verify authorization"]))
            body.append("if req.Auth.Error != nil {")
            body.extend(
                indent(
                    [
                        'setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)'
                    ],
                    1,
                )
            )
            body.append("} else {")
            body.extend(indent(call_block, 1))
            body.append("}")
        else:
            body.extend(call_block)
        body.append("")

        # Write the first populated response discovered and finish the handler
        body.extend(comment(["Write response to client"]))
        responses = [r for r in operation.responses]
        if ensure_500 and 500 not in {r.code for r in operation.responses}:
            responses.append(
                operations.Response(
                    code=500,
                    description="",
                    json_body_type=f"{api_package}.InternalServerErrorBody",
                )
            )
        for response in responses:
            body.append(f"if response.{response.response_set_field} != nil {{")
            body.extend(
                indent(
                    [
                        f"{api_package}.WriteJSON(w, {response.code}, response.{response.response_set_field})"
                    ],
                    1,
                )
            )
            body.extend(indent(["return"], 1))
            body.append("}")
        body.append(
            '%s.WriteJSON(w, 500, %s.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})'
            % (api_package, api_package)
        )

        lines.extend(indent(body, 1))

        lines.append("}")
        lines.append("")

    if need_auth_error:
        lines.append(
            "func setAuthError(ctx context.Context, authErr error, resp401 **ErrorResponse, resp403 **ErrorResponse, resp500 **api.InternalServerErrorBody) {"
        )

        body: list[str] = []

        body.append("switch stacktrace.GetCode(authErr) {")
        body.append("case dsserr.Unauthenticated:")
        body.extend(
            indent(
                [
                    '*resp401 = &ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authentication failed"))}'
                ],
                1,
            )
        )
        body.append("case dsserr.PermissionDenied:")
        body.extend(
            indent(
                [
                    '*resp403 = &ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authorization failed"))}'
                ],
                1,
            )
        )
        body.append("default:")
        body.extend(indent(["if authErr == nil {"], 1))
        body.extend(indent(['authErr = stacktrace.NewError("Unknown error")'], 2))
        body.extend(indent(["}"], 1))
        body.extend(
            indent(
                [
                    '*resp500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Could not perform authorization"))}'
                ],
                1,
            )
        )
        body.append("}")

        lines.extend(indent(body, 1))
        lines.append("}")
        lines.append("")

        imports.add("github.com/interuss/stacktrace")
        imports.add('dsserr "github.com/interuss/dss/pkg/errors"')

    if lines:
        lines.pop()

    return lines, imports


def routing(api: apis.API, api_package: str) -> list[str]:
    """Generate Go code to create an APIRouter for the provided Implementation.

    :param api: API to have its operation routes rendered
    :param api_package: Name of root/common API package
    :return: Lines of Go code defining the contents of a function to create an APIRouter that routes to the appropriate methods in the provided Implementation
    """
    lines: list[str] = []
    lines.append(
        "router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*%s.Route, %d)}"
        % (api_package, len(api.operations))
    )
    lines.append("")
    first_assignment = True
    for i, operation in enumerate(api.operations):
        prefix = ("/" + api.path_prefix) if api.path_prefix else ""
        path_regex = prefix + re.sub(r"{([^}]*)}", r"(?P<\1>[^/]*)", operation.path)
        lines.append(
            'pattern {}= regexp.MustCompile("^{}$")'.format(
                ":" if first_assignment else "", path_regex
            )
        )
        lines.append(
            'router.Routes[%d] = &%s.Route{Method: %s, Pattern: pattern, Handler: router.%s, Name: "%s.%s", Path: "%s%s"}'
            % (
                i,
                api_package,
                operation.verb_const_name,
                operation.interface_name,
                api.package,
                operation.interface_name,
                prefix,
                operation.path,
            )
        )
        lines.append("")
        first_assignment = False
    lines.append("return router")
    return lines


def example_implementation(api: apis.API, implementation_name: str) -> list[str]:
    """Generate Go code for a dummy API Implementation and a main routine to run it.

    Note that this routine produces a starting point/example for implementation,
    but it would be unusual to run it again when updating the API interface.  In
    that case, the data types and interface definitions would be overwritten by
    the new content auto-generated from the updated API interface, but the
    functions initially generated by this routine would generally be manually
    updated rather than being re-auto-generated.

    :param api: API to have its operation routes rendered
    :return: Lines of Go code defining a concrete instance of the Implementation interface
    """
    lines: list[str] = []

    # Define a concrete instance of the Implementation interface
    lines.append("type %s struct {}" % implementation_name)
    lines.append("")
    for operation in api.operations:
        lines.append(
            "func (*{}) {}(ctx context.Context, req *{}) {} {{".format(
                implementation_name,
                operation.interface_name,
                api.package + "." + operation.request_type_name,
                api.package + "." + operation.response_type_name,
            )
        )

        body: list[str] = []
        body.append(
            "response := %s{}" % (api.package + "." + operation.response_type_name)
        )
        # body.append('response.Response500 = &InternalServerErrorBody{ErrorMessage: "Not yet implemented"}')
        response_type = operation.responses[0].json_body_type
        if not response_type:
            response_type = "EmptyResponseBody"
        body.append(
            "response.%s = &%s{}"
            % (
                operation.responses[0].response_set_field,
                api.package + "." + response_type,
            )
        )
        body.append("return response")
        lines.extend(indent(body, 1))

        lines.append("}")
        lines.append("")

    return lines


def example_router_defs(implementations: dict[str, str], api_package: str) -> list[str]:
    """Generate Go code for concrete example router & multi-router definitions.

    :param implementations: Relationship between API name and the name of the Go struct implementing the Implementation of that API
    :param api_package: Name of root/common API package
    :return: Lines of Go code for router definitions in a main function to use to handle HTTP requests
    """
    lines: list[str] = []

    for api_name, implementation in implementations.items():
        lines.append(
            "%sRouter := %s.MakeAPIRouter(&%s{}, &authorizer)"
            % (api_name, api_name, implementation)
        )
    router_list = ", ".join(
        f"&{api_name}Router" for api_name, _ in implementations.items()
    )
    lines.append(
        "multiRouter := %s.MultiRouter{Routers: []%s.PartialRouter{%s}}"
        % (api_package, api_package, router_list)
    )

    return lines
