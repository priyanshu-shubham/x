# Installation

## Quick install

::: code-group
```bash [macOS / Linux]
curl -fsSL https://raw.githubusercontent.com/priyanshu-shubham/x/main/install.sh | sh
```

```powershell [Windows]
irm https://raw.githubusercontent.com/priyanshu-shubham/x/main/install.ps1 | iex
```
:::

This downloads the latest release and installs it to `~/.local/bin`.

## From source

If you have Go installed:

```bash
go install github.com/priyanshu-shubham/x@latest
```

## Verify installation

```bash
x version
```

## Configure AI features (optional)

If you want to use AI features (plain English commands, LLM steps, agentic mode):

```bash
x configure
```

You can use:
- **Anthropic API key** - Get one at [console.anthropic.com](https://console.anthropic.com)
- **Google Cloud Vertex AI** - For enterprise deployments

::: info Not required for command runner
If you only want to use `x` as a project command runner (like npm scripts or Makefiles), you don't need an API key. The AI features are optional.
:::

## Troubleshooting

### Command not found

Make sure `~/.local/bin` is in your PATH:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Or for zsh:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### API errors

Run `x configure` to check your credentials. Make sure you have credits/quota remaining with your API provider.

## Upgrading

```bash
x upgrade
```

This downloads and installs the latest version.
