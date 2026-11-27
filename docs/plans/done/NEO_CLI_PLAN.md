# `neo-cli` Plan

## Goal

Add a `neo-cli` binary that exposes the tools the MCP server offers without the HTTP server or an
MCP client, exactly as `nano-mcp` does with `cmd/nano-cli`. `serve` runs the MCP server itself, so
`cmd/server` is replaced rather than kept alongside.

The command surface is:

```text
neo-cli help
neo-cli serve
neo-cli forge <prompt...>
neo-cli openai <prompt...>
neo-cli edit <image-url>
neo-cli convert <image-url>
neo-cli bgkill <image-url>
```

## Scope

In scope:

- `cmd/neo-cli` as the only entry point, replacing `cmd/server`.
- An `internal/cli` package for behavior shared across commands.
- One command per MCP tool, writing generated and edited images to local files.
- Unit and command-level tests, README and Dockerfile updates.

Out of scope:

- Changing the MCP tool surface or server behavior.
- Credential storage, `login`, and `logout`: unlike NanoGPT, the keys these tools need already come
  from the server's own environment, so the CLI reads them from there.
- Publishing prebuilt binaries, a TUI, or shell completion installation.

## Decisions

- `github.com/spf13/cobra`, as in `nano-mcp`, with `SilenceUsage` and `SilenceErrors` on the root so
  a runtime failure prints one line.
- Every command loads `internal/config`, which is where `SD_URL`, `OPENAI_PROVIDERS`,
  `IMAGE_URL_MAP`, and `OUTPUT_FORMAT` already live, so the CLI and the server agree on defaults.
- Images are written locally and never hosted: `--out` selects the directory (default `.`), files
  are written atomically with mode `0644`, and are named `neo-image-<UTC stamp>-<random>.<ext>`.
- Image inputs are resolved through `internal/resolve` with `resolve.New(nil, "", sources)`, so
  http(s) URLs, data URIs, and raw base64 work without file hosting; a mapped local address fails
  with the resolver's own "file storage is not configured" error.
- Output mirrors the MCP handlers: forge and openai print each saved path, edit, convert, and
  bgkill print the one path, and the openai command prints the reported cost.

## Design

### Package layout

```text
cmd/neo-cli/main.go          root command and process entry point
cmd/neo-cli/serve.go         MCP server command, logger, signal handling
cmd/neo-cli/forge.go         forge command
cmd/neo-cli/openai.go        openai command
cmd/neo-cli/edit.go          edit command
cmd/neo-cli/convert.go       convert command
cmd/neo-cli/bgkill.go        bgkill command
internal/cli/deps.go         dependency injection, config, clients, resolver
internal/cli/asset.go        local image writing
internal/cli/output.go       shared formatting helpers
internal/cli/flags.go        optional flag helpers
```

`cmd/neo-cli` may import `internal/{config,forge,openai,edit,format,resolve,sourcemap,server}`.
It must not import `internal/mcp`, which stays the presentation layer for MCP content.

### Dependency injection

`internal/cli.Deps` carries `In`, `Out`, `Err`, `Config config.Config`, `Forge *forge.Client`,
`OpenAI *openai.Client`, and `Resolver *resolve.Client`. Nil clients are built from `Config`, and
`WithDefaults` fills the streams, so tests can inject a client pointed at an `httptest` server and
capture output in a buffer.

`internal/cli.Assets` provides:

- `WriteImage(dir, data []byte, mediaType string) (string, error)`, which picks the extension from
  the media type, creates the directory, writes through a temporary file, and renames it into
  place.

### Commands

`serve` is the old `cmd/server/main.go` moved behind Cobra: load config, build the logger, run
`internal/server`, and shut down gracefully on `SIGINT`/`SIGTERM` with the same 15 second window.

`forge <prompt...>` mirrors the `forge` tool: `--model` and `--forge-preset` are required,
`--vae-and-text-model` is repeatable, and the rest of the tool's inputs map one to one
(`--negative-prompt`, `--sampler-name`, `--scheduler`, `--steps`, `--width`, `--height`,
`--cfg-scale`, `--seed`, `--init-image-url`, `--denoising-strength`, `--enable-hr`, `--hr-scale`,
`--hr-upscaler`, `--hr-second-pass-steps`, `--hr-cfg`). `--format` and `--out` select the output.

`openai <prompt...>` mirrors the `openai` tool: `--provider`, `--model`, `--n`, `--size`, `--image`,
`--images` (repeatable), `--mask`, `--seed`, `--strength`, `--guidance-scale`,
`--num-inference-steps`, `--format`, `--out`. It builds its providers from `OPENAI_PROVIDERS` and
fails with a clear message when none are configured. A returned URL instead of base64 is downloaded
before it is written.

`edit <image-url>` mirrors the `edit` tool: the input is trimmed to its content, then `--square`,
`--circle`, and `--padding` refine it, with the same default padding rule.

`convert <image-url>` requires `--format` and fails when the source is already in it, as `Reformat`
does.

`bgkill <image-url>` takes `--model` (default `General`), `--full-mode`, `--format`, and `--out`.

Prompts are joined from their arguments, an empty prompt is rejected before any request, and a
prompt beginning with `-` uses `--` as usual.

### Errors and exit codes

Success exits `0`; usage and upstream failures exit `1` with one line on stderr. Commands never
write a partial file: the asset is written only after the operation succeeded.

## Implementation Units

### Unit 1: Scaffold, `serve`, and shared helpers

Add Cobra, `cmd/neo-cli/main.go`, `cmd/neo-cli/serve.go`, `internal/cli`, and move the entry point
out of `cmd/server`. Update the Dockerfile to build `cmd/neo-cli`, copy it to `/neo-cli`, and use
`["/neo-cli", "serve"]` as the entry point.

Acceptance criteria:

- [x] `go build ./cmd/neo-cli` produces a binary whose `neo-cli help` exits 0 with the tree above.
- [x] `cmd/server` no longer exists and nothing in the repo references it.
- [x] The Dockerfile builds `./cmd/neo-cli` and its entry point is `neo-cli serve`.
- [x] `neo-cli serve` loads `.env`/environment config and shuts down gracefully on `SIGINT`.
- [x] `internal/cli` has tests for config defaults, injected clients, and asset writing (extension,
      atomic rename, unknown media type, missing directory creation).

### Unit 2: `forge` and `openai`

Acceptance criteria:

- [x] `forge` sends the request `GenerateRequest` the MCP tool builds (checkpoint, preset, prompt,
      negative prompt, sampler, scheduler, steps, size, CFG, seed, HR fields) to an `httptest`
      Forge stub and writes each returned image to `--out`.
- [x] `forge` fetches `--init-image-url` through the resolver and sets `InitImage`, which selects
      img2img.
- [x] `forge` rejects a missing `--model`, a missing `--forge-preset`, and an empty prompt before
      any request.
- [x] `openai` posts the `ImageRequest` the MCP tool builds and writes inline and URL-returned
      images, printing each path and the reported cost.
- [x] `openai` resolves `--image`, `--images`, and `--mask` into data URLs with their media types.
- [x] `openai` with no `OPENAI_PROVIDERS` fails with a clear message and sends no request.
- [x] Tests cover success, an upstream HTTP error, and each validation failure.

### Unit 3: `edit`, `convert`, and `bgkill`

Acceptance criteria:

- [x] `edit` trims to content, honours `--square`, `--circle`, and `--padding`, re-encodes to the
      requested format, and writes one file.
- [x] `convert` re-encodes to `--format`, and refuses a source already in that format with the
      `Reformat` error.
- [x] `bgkill` posts the base64 payload the tool posts, decodes the foreground, and writes it.
- [x] Each command surfaces a fetch failure, an undecodable image, and an unsupported output format
      without writing a file.
- [x] Tests cover success and those failure paths.

### Unit 4: Documentation

Acceptance criteria:

- [x] `README.md` documents `go install github.com/wishmatic/neo-mcp/cmd/neo-cli@latest`, the
      commands above, that credentials come from `SD_URL` and `OPENAI_PROVIDERS`, and that images
      are written locally.
- [x] The Docker and local-development examples use `neo-cli serve`.
- [x] `go build ./...`, `go vet ./...`, and `go test ./... -race -count=1` pass.

## Verification Commands

```sh
go build ./...
go vet ./...
go test ./... -race -count=1
go build -o /tmp/neo-cli ./cmd/neo-cli
/tmp/neo-cli help
/tmp/neo-cli serve --help
/tmp/neo-cli forge --help
```

Human check: run `forge` against a local Forge instance and open the written image; run `openai`
with a real `OPENAI_PROVIDERS` entry and confirm the cost is printed.
