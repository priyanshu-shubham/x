# Config Files

`x` looks for configuration in two places: a global config and project-specific configs.

## Global config

Your global commands live in:

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/x/commands.yaml` |
| Linux | `~/.config/x/commands.yaml` |
| Windows | `%LOCALAPPDATA%\x\commands.yaml` |

Edit with:

```bash
x commands
```

This opens the file in your default editor.

Commands defined here are available everywhere on your system.

## Project config

Create `xcommands.yaml` in any directory. Commands defined here are available when you're in that directory or any subdirectory.

```
myproject/
├── xcommands.yaml    # Available in myproject/ and below
├── src/
│   └── ...
└── tests/
    └── ...
```

## Config merging

All configs are merged together. If you're in `/projects/myapp/src/`, x loads:

1. Global `commands.yaml` (base)
2. `/projects/xcommands.yaml` (if it exists)
3. `/projects/myapp/xcommands.yaml` (if it exists)

**Later configs override earlier ones.** If both global and local define `build`, the local one wins.

## Seeing where commands come from

Run `x --help` to see the source of each command:

```
Built-in commands:
  configure   Configure API credentials
  commands    Edit global commands
  usage       Show token usage
  upgrade     Upgrade to latest version
  version     Show version

Commands (global):
  shell       Generate shell commands from natural language (default)
  new         Create a new command with AI assistance

Commands (myapp):
  build       Build the project
  test        Run tests
```

## File format

Both global and local configs use the same YAML format:

```yaml
# Optional: set a default command
default: shell

# Define commands
build:
  description: Build the project
  steps:
    - exec:
        command: go build -o app .

test:
  description: Run tests
  args:
    - name: filter
      description: Test filter pattern
  steps:
    - exec:
        command: go test -run {{args.filter}} ./...
```

## Default command

The `default` field specifies which command runs when no match is found:

```yaml
default: shell

shell:
  description: Generate shell commands
  args:
    - name: prompt
      rest: true
  steps:
    - llm:
        system: Generate a shell command for the user's request.
        prompt: "{{args.prompt}}"
```

Now `x find large files` runs the `shell` command with "find large files" as the prompt.

## Auto-created config

If the global config doesn't exist, it's created automatically with default commands (`shell` and `new`) the first time you run `x`.

## Config tips

- **Start global, then specialize** - Put commonly used commands in global config, project-specific ones in local
- **Commit project configs** - Share `xcommands.yaml` with your team via version control
- **Use descriptive names** - Good: `build-docker`, `test-integration`. Bad: `bd`, `ti`
- **Add descriptions** - They show up in `x --help` and make commands self-documenting

## Debugging config issues

### Syntax errors

YAML is picky about indentation. Use 2 spaces, not tabs.

```bash
DEBUG=1 x --help
```

Shows parsing errors if any.

### Wrong command running

Check which config is being used:

```bash
x --help
```

Look at the source in parentheses: `(global)` vs `(projectname)`.

### Reset to defaults

Delete the global config and run any command - it'll be recreated:

```bash
rm ~/.config/x/commands.yaml  # Linux
x --help  # Recreates with defaults
```
