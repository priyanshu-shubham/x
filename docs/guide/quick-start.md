# Quick Start

Get up and running in 5 minutes. If you haven't read [What is x?](/guide/), start there to understand the concepts.

## 1. Install

::: code-group
```bash [macOS / Linux]
curl -fsSL https://raw.githubusercontent.com/priyanshu-shubham/x/main/install.sh | sh
```

```powershell [Windows]
irm https://raw.githubusercontent.com/priyanshu-shubham/x/main/install.ps1 | iex
```

```bash [From source]
go install github.com/priyanshu-shubham/x@latest
```
:::

Verify:

```bash
x version
```

## 2. Try the pre-installed commands

### Option A: With AI (needs API key)

Configure your API key:

```bash
x configure
```

Then try the `shell` command:

```bash
x find all files larger than 100mb
```

You'll see the generated command and can choose to run it.

### Option B: Without AI

Create `xcommands.yaml` in any folder:

```yaml
hello:
  description: Say hello
  steps:
    - exec:
        command: echo "Hello from x!"
```

Run it:

```bash
x hello
```

No API key needed for `exec` steps.

## 3. Create a project command

In your project folder, create `xcommands.yaml`:

```yaml
build:
  description: Build the project
  steps:
    - exec:
        command: npm run build

test:
  description: Run tests
  steps:
    - exec:
        command: npm test

dev:
  description: Start dev server
  steps:
    - exec:
        command: npm run dev
```

Now use them:

```bash
x build
x test
x dev
x --help    # See all commands
```

## 4. Create a global command

Open the global config:

```bash
x commands
```

Add a command you want everywhere:

```yaml
ip:
  description: Show my public IP
  steps:
    - exec:
        command: curl -s ifconfig.me
```

Save and close. Now `x ip` works in any directory.

## 5. Learn about the built-in commands

The `shell` and `new` commands are built into the binary and always up-to-date:

```bash
x shell --help    # See description and usage
x --help          # See all commands and their sources
```

You can override them in your config if you want custom behavior, but the built-in versions are automatically updated when you upgrade `x`.

## Useful commands

| Command | What it does |
|---------|--------------|
| `x --help` | List all available commands |
| `x <cmd> --help` | Help for a specific command |
| `x commands` | Edit global commands |
| `x configure` | Set up API key |
| `x upgrade` | Update to latest version |
| `x usage` | Check API usage and costs |

## Debug and dry run

See what's happening:

```bash
DEBUG=1 x build
```

Preview without executing:

```bash
DRYRUN=1 x "delete all temp files"
```

## Next steps

- [Your First Command](/guide/first-command) - Build a complete command with arguments and cross-platform support
- [Building a Python Fixer](/guide/python-fixer) - Create an AI agent that fixes code
- [Examples](/examples/) - Ready-to-use commands to copy
