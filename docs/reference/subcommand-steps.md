# subcommand Steps

Call another command as a step. Great for composing complex workflows from simpler building blocks.

## Basic usage

```yaml
# Define a simple command
build:
  description: Build the project
  steps:
    - exec:
        command: go build -o app .

# Call it from another command
release:
  description: Build and tag a release
  args:
    - name: version
  steps:
    - subcommand:
        name: build
    - exec:
        command: git tag {{args.version}}
```

## Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `name` | string | required | Name of the command to call |
| `args` | list | `[]` | Arguments to pass |
| `silent` | boolean | `false` | Don't print "Running subcommand:" message |

## Passing arguments

Pass arguments as a list:

```yaml
review:
  description: Review a file
  args:
    - name: file
  steps:
    - exec:
        command: cat {{args.file}}
        silent: true
    - llm:
        system: Review this code.
        prompt: "{{output}}"

review-all:
  description: Review multiple files
  steps:
    - subcommand:
        name: review
        args:
          - "src/main.go"
    - subcommand:
        name: review
        args:
          - "src/config.go"
```

## Variable interpolation in args

Arguments support all template variables:

```yaml
review-and-fix:
  description: Review then fix
  args:
    - name: file
  steps:
    - id: review
      subcommand:
        name: review
        args:
          - "{{args.file}}"
    - subcommand:
        name: fix
        args:
          - "Fix these issues in {{args.file}}: {{steps.review.output}}"
```

## Capturing output

Subcommand output is captured automatically and available as <code v-pre>{{output}}</code> or via named steps:

```yaml
pipeline:
  steps:
    - id: step1
      subcommand:
        name: generate-data
    - subcommand:
        name: process-data
        args:
          - "{{steps.step1.output}}"
```

## Silent mode

By default, calling a command via subcommand step prints a message:

```
Running subcommand: build
```

Use `silent: true` to suppress this:

```yaml
steps:
  - subcommand:
      name: build
      silent: true
```

## Example: Composable workflow

Build complex workflows from simple, reusable commands:

```yaml
# Basic building blocks
lint:
  description: Run linter
  steps:
    - exec:
        command: golangci-lint run

test:
  description: Run tests
  steps:
    - exec:
        command: go test ./...

build:
  description: Build binary
  steps:
    - exec:
        command: go build -o app .

# Composed workflows
check:
  description: Lint and test
  steps:
    - subcommand:
        name: lint
    - subcommand:
        name: test

release:
  description: Full release pipeline
  args:
    - name: version
  steps:
    - subcommand:
        name: check
    - subcommand:
        name: build
    - exec:
        command: git tag {{args.version}} && git push origin {{args.version}}
        confirm: true
```

Now:
- `x lint` - Just lint
- `x test` - Just test
- `x check` - Lint and test
- `x release v1.0.0` - Full pipeline

## Recursive calls

Commands can call other commands that call other commands. There's no built-in recursion limit, so be careful not to create infinite loops.
