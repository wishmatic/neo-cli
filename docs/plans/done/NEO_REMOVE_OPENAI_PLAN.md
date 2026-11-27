# Remove the `openai` Backend Plan

## Goal

Delete the OpenAI-compatible image backend from neo-mcp, in both the MCP tool and the CLI. It predates
the Forge tool, has no users, and every path it owns is a second configuration surface to keep
working.

## Scope

In scope:

- The `openai` MCP tool, its request builders, and its schema.
- `internal/openai` and `OPENAI_PROVIDERS` in `internal/config`.
- Server wiring that builds the client and enables the tool.
- `neo-cli gen openai` and `internal/cli`'s OpenAI dependency.
- README, `docs/PROMPT.md`, `.env.example`, and the AGENTS architecture diagram.

Out of scope:

- The `forge`, `edit`, `convert`, and `bgkill` tools and commands.
- `OUTPUT_FORMAT`, `IMAGE_URL_MAP`, and the rest of the image pipeline.

## Design

Deletion only: no behaviour is replaced. `forge` becomes the sole generator on both surfaces, so
`neo-cli gen image` and the `forge` tool are what remains.

## Implementation Units

### Unit 1: MCP tool and server wiring

Delete `internal/mcp/openai.go`, `internal/mcp/openai_images.go`, `internal/mcp/openai_test.go`, the
OpenAI helpers in `internal/mcp/backends_test.go`, the whole `internal/openai` package, and
`Clients.OpenAI`. Remove `registerOpenAI` from `registerTools`, and stop building the client from
`OPENAI_PROVIDERS` in `internal/server`.

Acceptance criteria:

- [x] The MCP server lists `forge`, `edit`, `convert`, and `bgkill` and no `openai` tool.
- [x] A config with `OPENAI_PROVIDERS` set is unaffected, and the variable no longer exists.
- [x] `internal/openai` no longer exists and nothing imports it.
- [x] `go build ./...` and `go test ./... -count=1` pass.

### Unit 2: CLI

Delete `cmd/neo-cli/gen_openai.go`, its tests, `cli.Deps.OpenAI`, `cli.OpenAIClient`, and the
`newGenOpenAICommand` registration.

Acceptance criteria:

- [x] `neo-cli gen` lists only `image`, and `neo-cli gen openai` fails as an unknown command.
- [x] `internal/cli` no longer references OpenAI.
- [x] `go build ./...` and `go test ./... -count=1` pass.

### Unit 3: Documentation

Update `README.md` (features, CLI examples, agent model knowledge), `docs/PROMPT.md` (the
openai-model section and the credits note), `.env.example`, and the AGENTS architecture diagram.

Acceptance criteria:

- [x] No source, doc, or config file mentions `OPENAI_PROVIDERS`, `gen openai`, or an
      OpenAI-compatible endpoint.
- [x] The AGENTS diagram has no `openai` node or edges.
- [x] `go build ./...`, `go vet ./...`, and `go test ./... -race -count=1` pass, and
      `neo-cli help`, `neo-cli gen --help`, and `neo-cli gen image --help` exit 0.

## Verification Commands

```sh
go build ./...
go vet ./...
go test ./... -race -count=1
grep -ri openai --include=*.go --include=*.md --include=.env.example .
```

Human check: with a Forge instance running, `neo-cli gen image "a red fox" --model <checkpoint>
--forge-preset <preset> --out ./images` still writes an image.
