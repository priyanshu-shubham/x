# Variables

Use <code v-pre>{{variable}}</code> syntax to insert dynamic values anywhere in your commands.

## Argument variables

Access user-provided arguments with <code v-pre>{{args.name}}</code>:

```yaml
greet:
  args:
    - name: name
      description: Name to greet
  steps:
    - exec:
        command: echo "Hello, {{args.name}}!"
```

```bash
x greet World
# Output: Hello, World!
```

### Rest arguments

Use `rest: true` to capture all remaining arguments as a single string:

```yaml
search:
  args:
    - name: query
      rest: true
  steps:
    - exec:
        command: grep -r "{{args.query}}" .
```

```bash
x search hello world
# {{args.query}} = "hello world"
```

## Step output variables

### Previous step output

Use <code v-pre>{{output}}</code> to access the previous step's output:

```yaml
steps:
  - exec:
      command: cat README.md
      silent: true
  - llm:
      prompt: "Summarize: {{output}}"
```

### JSON field access

If a step outputs JSON, access individual fields with <code v-pre>{{output.field}}</code>:

```yaml
steps:
  - llm:
      system: |
        Return JSON: {"command": "...", "summary": "..."}
      prompt: "{{args.task}}"
      silent: true
  - exec:
      command: "{{output.command}}"
      summary: "{{output.summary}}"
      confirm: true
```

This also works with named steps: <code v-pre>{{steps.id.field}}</code>

::: warning
If the output is not valid JSON, accessing a field will cause an error. Use <code v-pre>{{output}}</code> (without a field) for raw output.
:::

### Named step output

Give a step an `id` to reference its output specifically:

```yaml
steps:
  - id: readme
    exec:
      command: cat README.md
      silent: true
  - id: changelog
    exec:
      command: cat CHANGELOG.md
      silent: true
  - llm:
      prompt: |
        README:
        {{steps.readme.output}}

        CHANGELOG:
        {{steps.changelog.output}}
```

## Environment variables

<div v-pre>

| Variable | Description | Example |
|----------|-------------|---------|
| `{{directory}}` | Current working directory | `/home/user/project` |
| `{{os}}` | Operating system | `linux`, `darwin`, `windows` |
| `{{arch}}` | CPU architecture | `amd64`, `arm64` |
| `{{shell}}` | User's shell | `/bin/bash`, `/bin/zsh` |
| `{{user}}` | Username | `john` |

</div>

Example:

```yaml
info:
  steps:
    - exec:
        command: echo "OS: {{os}}, Arch: {{arch}}, Shell: {{shell}}"
```

## Date and time variables

<div v-pre>

| Variable | Description | Example |
|----------|-------------|---------|
| `{{date}}` | Current date | `2024-01-15` |
| `{{time}}` | Current time | `14:30:05` |
| `{{datetime}}` | Date and time | `2024-01-15 14:30:05` |

</div>

Example:

```yaml
log:
  args:
    - name: message
      rest: true
  steps:
    - exec:
        command: echo "[{{datetime}}] {{args.message}}" >> log.txt
```

## Using in different contexts

Variables work everywhere:

### In exec commands

```yaml
steps:
  - exec:
      command: echo "Running on {{os}} at {{datetime}}"
```

### In LLM prompts

```yaml
steps:
  - llm:
      system: |
        You are helping a user on {{os}}.
        Current directory: {{directory}}
      prompt: "{{args.question}}"
```

### In OS-specific overrides

```yaml
steps:
  - exec:
      command: cat {{args.file}}
      windows: type {{args.file}}
```

### In subcommand args

```yaml
steps:
  - subcommand:
      name: process
      args:
        - "{{args.file}}"
        - "{{output}}"
```

## Variable reference

<div v-pre>

| Variable | Source | Description |
|----------|--------|-------------|
| `{{args.name}}` | User input | Named argument value |
| `{{output}}` | Previous step | Raw output from the previous step |
| `{{output.field}}` | Previous step | JSON field from the previous step (errors if not JSON) |
| `{{steps.id.output}}` | Named step | Raw output from a specific step |
| `{{steps.id.field}}` | Named step | JSON field from a specific step |
| `{{directory}}` | Runtime | Current working directory |
| `{{os}}` | Runtime | Operating system (`linux`, `darwin`, `windows`) |
| `{{arch}}` | Runtime | CPU architecture |
| `{{shell}}` | Runtime | User's default shell |
| `{{user}}` | Runtime | Current username |
| `{{date}}` | Runtime | Current date (YYYY-MM-DD) |
| `{{time}}` | Runtime | Current time (HH:MM:SS) |
| `{{datetime}}` | Runtime | Current date and time |

</div>
