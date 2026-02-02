# x

> Describe what you want in plain English, get the shell command. Build custom AI shortcuts too.

Instead of googling "how to find files modified in the last 24 hours", just ask:

```
x find files modified in the last day
```

It shows you the command and asks if you want to run it.

Create your own subcommands with custom prompts:

```
x explain "what is a mutex"
x commit-msg "added JWT authentication"
x translate-to-spanish "Hello, how are you?"
```

## Installation

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/priyanshu-shubham/x/main/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/priyanshu-shubham/x/main/install.ps1 | iex
```

### Windows (Git Bash / WSL)

Use the same command as macOS / Linux.

### From source

If you have Go installed:

```bash
go install github.com/priyanshu-shubham/x@latest
```

Or clone and build:

```bash
git clone https://github.com/priyanshu-shubham/x.git
cd x
go build -o x .
```

## Setup

After installing, configure your API credentials:

```
x configure
```

You can use either an Anthropic API key or Google Cloud Vertex AI.

That's it. Start using it:

```
x list all png files larger than 1mb
```

## How it works

You type what you want in plain English. Claude generates the appropriate shell command for your OS. You see the command and can choose to run it or not (default is yes, just hit Enter).

```
$ x count lines of code in all python files

  find . -name "*.py" -exec wc -l {} + | tail -1

Run this command? [Y/n]:
```

Works on macOS, Linux, and Windows. It'll generate the right commands for your platform.

## Custom subcommands

The real power is creating your own subcommands with custom prompts.

Run `x subcommands` to open the config file, then add something like:

```yaml
explain:
  prompt: |
    Explain the following concept in simple terms.
    Be concise and use examples where helpful.

commit-msg:
  prompt: |
    Generate a concise git commit message for these changes.
    Output only the message, nothing else.
```

Now you can use them:

```
$ x explain "what is a mutex"
A mutex (mutual exclusion) is like a bathroom key at a coffee shop...

$ x commit-msg "added user authentication with JWT tokens"
Add JWT-based user authentication
```

### Template variables

You can use these in your prompts and they'll be replaced at runtime:

- `{{time}}` - Current time (HH:MM:SS)
- `{{date}}` - Current date (YYYY-MM-DD)
- `{{datetime}}` - Both
- `{{directory}}` - Current working directory
- `{{os}}` - Operating system
- `{{arch}}` - Architecture
- `{{shell}}` - Your shell
- `{{user}}` - Your username

Example:

```yaml
journal:
  prompt: |
    You're helping me write a daily journal entry.
    Today is {{date}} and the time is {{time}}.
    I'm currently in {{directory}}.
    Help me reflect on what I describe.
```

## Upgrade

Update to the latest version:

```
$ x upgrade
Current version: v0.1.0
Checking for updates...
Latest version: v0.2.0
Downloading x_darwin_arm64.tar.gz...
Upgraded to v0.2.0
```

Check current version with `x version`.

## Usage stats

Track how many tokens you've used and the estimated cost:

```
$ x usage
Token Usage
-----------
Input tokens:          1250
Output tokens:         380
Total tokens:          1630
Requests:              12

Estimated Cost
--------------
Input cost:            $0.0038
Output cost:           $0.0057
Total cost:            $0.0095
```

Pricing is fetched dynamically from [LiteLLM's pricing data](https://github.com/BerriAI/litellm).

## Config location

Everything lives in your system's config directory:

- macOS: `~/Library/Application Support/x/`
- Linux: `~/.config/x/`
- Windows: `%LOCALAPPDATA%\x\`

Files:
- `config.json` - Your API credentials
- `subcommands.yaml` - Your custom subcommands
- `usage.json` - Token usage stats

## Tips

- The default command (without a subcommand) always generates shell commands and asks before running
- Custom subcommands just output text directly - useful for anything that isn't a shell command
- If you mess up the subcommands.yaml, just delete it and run `x subcommands` to get a fresh template

## Releasing (for maintainers)

To create a new release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will automatically build binaries for all platforms and create a release.
