# What is x?

`x` is a command runner for your projects - like npm scripts or Makefiles, but simpler. You can also add AI to your commands - call Claude to generate text, analyze files, or run commands autonomously. Mix shell scripts and AI however you want.

## The basics

You define **commands** in a YAML file, then run them with `x <name>`:

```yaml
# xcommands.yaml
build:
  description: Build the project
  steps:
    - exec:
        command: go build -o app .

test:
  description: Run tests
  steps:
    - exec:
        command: go test ./...
```

```bash
x build    # Runs: go build -o app .
x test     # Runs: go test ./...
```

That's the core idea. No AI needed - just simple command shortcuts.

## What's a command?

A command is a named task with one or more steps. Each step does something:

- **[exec](/reference/exec-steps)** - Run a shell command
- **[llm](/reference/llm-steps)** - Call Claude (AI)
- **[agentic](/reference/agentic-steps)** - Let Claude run commands autonomously (AI)
- **[subcommand](/reference/subcommand-steps)** - Call another command

Most commands just use `exec`. AI steps are optional.

## Where commands live

Commands can be defined in two places:

### Project commands

Create `xcommands.yaml` in your project folder. Commands here only work in that directory (and subdirectories).

```
myproject/
├── xcommands.yaml    # x build, x test work here
├── src/
│   └── ...           # and here
└── README.md
```

### Global commands

Commands you want everywhere go in your global config:

- **macOS:** `~/Library/Application Support/x/commands.yaml`
- **Linux:** `~/.config/x/commands.yaml`
- **Windows:** `%LOCALAPPDATA%\x\commands.yaml`

Run `x commands` to open it in your editor.

## Built-in commands

Two commands are built into the binary and always available:

**shell** - Generate shell commands from plain English (with safety info):
```bash
x find files larger than 100mb
```

**new** - Create new commands with AI:
```bash
x new "a command that compresses all images"
```

These are automatically updated when you upgrade `x`. You can override them in your config if you want custom behavior.

## The default command

The global config has a `default` setting:

```yaml
default: shell
```

When you run `x something` and "something" doesn't match any command name, it uses the default. That's why `x find large files` works - it passes "find large files" to the `shell` command.

## AI is optional

`x` works in two modes:

**Without AI (no API key needed):**
- Define commands with `exec` steps
- Run shell commands
- Works like a simple task runner

**With AI (requires API key):**
- Use the `shell` command to generate commands from English
- Add `llm` steps to call Claude
- Add `agentic` steps for autonomous workflows
- Use `new` to create commands with AI

Configure AI with `x configure`. Skip it if you only need the command runner.

## How commands are found

When you run `x build`:

1. Check **CLI commands** (`configure`, `commands`, `usage`, `upgrade`, `version`)
2. Check **built-in commands** (`shell`, `new` - embedded in the binary)
3. Check **global config** (`~/.config/x/commands.yaml`)
4. Check **project configs** (all `xcommands.yaml` files from root to current directory)
5. If no match, use the **default command** (if set)

Later configs override earlier ones. A project's `build` command overrides a global `build`. You can also override built-in commands in your config.

Run `x --help` to see all available commands and where they come from.

## Next: Get started

Ready to try it?

- [Quick Start](/guide/quick-start) - Install and run your first commands in 5 minutes
- [Your First Command](/guide/first-command) - Tutorial: build a complete command
- [Building a Python Fixer](/guide/python-fixer) - Tutorial: create an AI agent
