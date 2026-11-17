# AI Agent Collaboration Guide — `dss`

This document provides context and guidelines for AI agents working in the `dss` repository.

## User Documentation (`docs/`)

- **Consult Design Guidelines First**: When making any changes to user documentation within the `docs/` folder, always consult [`docs/documentation_design.md`](./docs/documentation_design.md) first to ensure changes align with the target user profile, terminology definitions, procedural structure, and user journeys.

## Development & Tooling Tips

- **CLI Flag Changes**: When adding, renaming, or removing CLI flags in `cmds/`, consult [`DEVELOPMENT.md`](./DEVELOPMENT.md#changing-command-line-flags-or-commands) for the full list of files to update across `docs/`, Helm, Tanka, Terraform, and `test/`.
- **OpenAPI Go Server Codegen (`interfaces/openapi-to-go-server/`)**: Modifying generator scripts or templates affects three generated trees (`pkg/api`, `interfaces/openapi-to-go-server/example`, and `cmds/dummy-oauth/api`, normally regenerated via `make apis`, which CI verifies produces zero `*.gen.go` diff). When adding a new protected endpoint or scope to `pkg/api`, also ensure its required OAuth scope is included in `ALL_SCOPES` in [`test/security/test.py`](./test/security/test.py) (which scrapes `Path:` fields from `pkg/api/**/*.go`).

## Continuous Improvement of this Guide

- **Pay It Forward**: As an AI agent, your ability to acquire context quickly is critical. If, during your work, you find yourself spending significant time overcoming a misunderstanding or discovering a hidden project nuance, **please proactively update this `AGENTS.md` file**. Add concise tips or warnings to help future agents (including yourself) avoid the same friction, alongside completing your core task.
- **Correcting Mistakes**: Documentation can become outdated or contain errors over time. If your core work reveals that information in this `AGENTS.md` file is mistaken or missing critical context, please take the initiative to fix those errors while completing your primary objectives!

