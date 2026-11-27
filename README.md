# Neo MCP

<img src="docs/images/logo.webp" alt="Neo MCP Logo" width="128">

A thin MCP wrapper exposing image capabilities to LLM agents.

## Features

Note that the only version of Forge we support is
[Haoming02's fork](https://github.com/Haoming02/sd-webui-forge-classic/tree/neo#stable-diffusion-webui-forge---neo).

- Core feature: `forge` generates images against your local Forge Neo instance.
  - It takes an optional `init_image_url`, and transforms that image instead of generating from
    scratch when it is given.
  - It is the tool with the Forge-only options: presets, VAE/text encoders, and hi-res fix.
- Not just PNG; WebP output by default, and JPEG and JXL are supported too!
  - Adjust default output with `OUTPUT_FORMAT`.
- Every call to an image tool stores the image and returns its URL as text; pass `inline: true` to
  attach the image as an MCP image block as well, such that a vision-capable agent can see the
  returned result.
  - The attached copy is shrunk to a size budget the caller can raise or lower per call with
    `inline_max_edge` and `inline_max_bytes`; the stored image at its URL keeps its own size.
  - This also means that this isn't just an MCP; it also is a simple image hosting service!
- `bgkill` removes the background from an image via the
  [`sd-webui-birefnet`](https://github.com/dimitribarbot/sd-webui-birefnet) extension for Forge.
- `edit` trims an image to the bounding box of its visible content, optionally padding, centering or
  even circularising it! Great for logos and avatars. This runs locally and does not require Forge.
- `convert` re-encodes an image as `png`, `jpeg`, `jxl`, or `webp`. Also runs locally!
- Private networking support; input URLs can be mapped before they are fetched, even to a local
  directory seen by the container.

## Usage

### CLI

```sh
go install github.com/wishmatic/neo-mcp/cmd/neo-cli@latest
```

The CLI exposes the same tools as the MCP server without an MCP client. It reads its settings from
the server's environment: `SD_URL` for Forge, `OUTPUT_FORMAT` for the default output type, and
`IMAGE_URL_MAP` for input addresses. Generated and edited images are written locally, so no file
hosting is involved; `--out` picks the directory.

```sh
neo-cli help
neo-cli gen image "a red fox in snow" --model sd.safetensors --forge-preset default --out ./images
neo-cli gen image "a red fox in snow" --init-image ./fox.png --model sd.safetensors --forge-preset default --out ./images
neo-cli edit square https://example.com/logo.png --out ./images
neo-cli edit circle ./logo.png --out ./images
neo-cli edit convert https://example.com/logo.png --format webp --out ./images
neo-cli edit bgkill https://example.com/photo.jpg --model General --out ./images
```

`gen` also answers to `generate`. `edit` has `trim`, `square`, `circle`, `convert`, and `bgkill`
subcommands, and every `<image>` they take is a local path or an address the resolver can read, such
as an http(s) URL, a data URI, or raw base64. `--init-image` takes a local file for img2img;
`--init-image-url` takes an address.

`neo-cli serve` runs the MCP server itself.

### Docker

Deploy as a Docker image:

```sh
docker run -d \
  -p 8080:8080 \
  -e API_KEY=change-me \
  -e PUBLIC_HOST=http://192.168.1.10:8080 \
  -e SD_URL=http://host.docker.internal:7860 \
  -v /mnt/user/appdata/neo-mcp:/data \
  ghcr.io/wishmatic/neo-mcp:latest
```

The MCP endpoint is served at `/mcp`.

`/data` holds the image store, so bind-mount a host directory there to keep files across container
replacements; the container runs as uid 65532, so that directory must be writable by it. The image's
entry point is `neo-cli serve`.

All other configuration is optional but strongly recommended; see [.env.example](.env.example).

### SD Web UI Forge Neo's API

You need to turn on the API for SD Web UI Forge Neo for this to work. Add `--api` to your Forge
launch command, then point `SD_URL` at your Forge instance's API.

No support will be provided for any issues related to your installation of Forge, nor the Forge API.

### Authentication

`API_KEY` is required on every `/mcp` request, sent as `Authorization: Bearer <API_KEY>`. Stored
images are served without authentication.

### Agent Model Knowledge

Your agent will need knowledge of VAE and text encoder models as well as available upscalers in
order to use them. Add these verbatim to your system prompt or a skill, along with checkpoints,
LoRA, and anything else it needs. No list is maintained here, so the agent supplies them.

### Stored Images

Please see [.env.example](.env.example) for a configuration. Put `PUBLIC_HOST` behind a reverse
proxy or CDN if you want caching; responses are immutable and cacheable.

### Extensions Support

Some extensions I use will be supported over time if they can be called via the Forge API.

Right now, this is only [`sd-webui-birefnet`](https://github.com/dimitribarbot/sd-webui-birefnet),
which provides the `bgkill` tool.

## Warnings

This MCP will be available and maintained so long as I use it, and is built for my own purposes.
Extending features via issue requests and PRs will be _considered_ but unless I find use out of it
myself, I probably won't work on those features.

This is also very bespoke to my use case. I recommend forking this and adjusting features to your
needs if it doesn't quite fit your own.

## Local Development

Modify `.env`; you can copy from `.env.example`. Then:

```sh
go run ./cmd/neo-cli serve
```

## License

Neo MCP is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).

This project is not affiliated with or endorsed by AUTOMATIC1111, Illyasviel, Haoming02, or any
other third-party services. Those names are trademarks of their respective owners and are used here
only to describe compatibility.
