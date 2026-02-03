# exec Steps

Run shell commands.

## Basic usage

```yaml
steps:
  - exec:
      command: echo "Hello, world!"
```

## Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `command` | string | required | The command to run |
| `windows` | string | - | Override command for Windows |
| `darwin` | string | - | Override command for macOS |
| `linux` | string | - | Override command for Linux |
| `silent` | boolean | `false` | Capture output without printing |
| `confirm` | boolean | `false` | Ask before running |

## OS-specific commands

Specify different commands for different operating systems:

```yaml
steps:
  - exec:
      command: cat file.txt        # Default (also used for Linux)
      windows: type file.txt       # Windows
      darwin: cat file.txt         # macOS (optional, uses default)
```

The OS is detected at runtime using Go's `runtime.GOOS`.

## Silent mode

When `silent: true`, the command runs but output isn't printed. The output is captured and available as <code v-pre>{{output}}</code> in the next step:

```yaml
steps:
  - exec:
      command: git diff --cached
      silent: true
  - llm:
      prompt: "Summarize these changes: {{output}}"
```

When `silent: false` (default), output streams to your terminal in real-time.

## Confirmation prompt

When `confirm: true`, the command is shown and you're asked before it runs:

```yaml
steps:
  - exec:
      command: rm -rf ./dist
      confirm: true
```

Output:
```
┃ rm -rf ./dist

Run this command? [Y/n]:
```

Useful for destructive commands.

## Variable interpolation

Use `{{}}` to insert values:

```yaml
build:
  args:
    - name: target
      description: Build target
  steps:
    - exec:
        command: go build -o {{args.target}} .
```

Available variables:
- <code v-pre>{{args.name}}</code> - User-provided arguments
- <code v-pre>{{output}}</code> - Output from previous step
- <code v-pre>{{steps.id.output}}</code> - Output from a named step
- <code v-pre>{{directory}}</code>, <code v-pre>{{os}}</code>, <code v-pre>{{shell}}</code>, etc. - See [Variables](/reference/variables) for full list

## Named steps

Give a step an `id` to reference its output later:

```yaml
steps:
  - id: version
    exec:
      command: git describe --tags
      silent: true
  - exec:
      command: echo "Building version {{steps.version.output}}"
```

## Long-running processes

If you run a long-lived process (like a server) as a non-final step, it will block the pipeline. Either:

1. Run it as the **last step**
2. Run it in the background with `&`:

```yaml
steps:
  - exec:
      command: ./server &
  - exec:
      command: curl http://localhost:8080/health
```

## Exit codes

If a command exits with a non-zero code, the pipeline stops and the error is reported.

To continue even if a command fails, append `|| true`:

```yaml
steps:
  - exec:
      command: rm -f ./maybe-exists.txt || true
  - exec:
      command: echo "Continuing..."
```
