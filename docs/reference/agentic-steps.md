# agentic Steps

Let Claude run commands autonomously to complete complex tasks.

## Basic usage

```yaml
steps:
  - agentic:
      system: You are a helpful assistant with shell access.
      prompt: "Find and delete all .tmp files"
      max_iterations: 10
      auto_execute: false
```

## Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `system` | string | - | System prompt with instructions |
| `prompt` | string | required | The task to complete |
| `max_iterations` | number | `10` | Max commands Claude can run |
| `auto_execute` | boolean | `false` | Run commands without confirmation |

## How it works

In agentic mode, Claude has access to a shell tool. It can:

1. Run shell commands to explore and understand
2. Make changes to files
3. Run more commands to verify
4. Signal completion when done

The loop continues until Claude calls the `complete` tool or reaches `max_iterations`.

## What the user sees

- **Claude's text responses** are displayed (rendered as markdown)
- **Commands** are shown with "Executing: ..." before running
- **Command output** is NOT shown to the user, only returned to Claude

This means Claude should summarize important results in its text responses.

## Auto-execute mode

When `auto_execute: false` (default), you approve each command:

```
Claude wants to run: cat src/main.py
[Y/n]:
```

When `auto_execute: true`, commands run without asking:

```yaml
steps:
  - agentic:
      system: Explore the codebase and summarize the architecture.
      prompt: "What does this project do?"
      auto_execute: true  # Be careful!
```

::: danger Use with caution
Auto-execute lets Claude run any command without confirmation. Only use this for read-only tasks or in controlled environments.
:::

## Max iterations

The `max_iterations` limit prevents runaway agents:

```yaml
steps:
  - agentic:
      prompt: "Fix all the bugs"
      max_iterations: 20  # Allow more iterations for complex tasks
```

If the limit is reached without completion, you'll see a warning:

```
⚠ Agent reached max iterations (20) without completing
```

## System prompt tips

Give Claude context about the environment and task:

```yaml
steps:
  - agentic:
      system: |
        You are a senior developer debugging an issue.

        Environment:
        - OS: {{os}}
        - Shell: {{shell}}
        - Directory: {{directory}}

        Available tools:
        - pytest for running tests
        - git for version control
        - Standard Unix tools (cat, grep, sed, etc.)

        Instructions:
        - Read files before making changes
        - Run tests after each change
        - Make minimal, targeted fixes
      prompt: "{{args.issue}}"
```

## Combining with other steps

Run commands before agentic mode to provide context:

```yaml
fix-tests:
  description: Fix failing tests
  steps:
    - id: tests
      exec:
        command: pytest --tb=short 2>&1 | head -50 || true
        silent: true
    - agentic:
        system: |
          Fix the failing tests.

          Current test output:
          {{steps.tests.output}}
        prompt: "Make the tests pass"
        max_iterations: 15
```

## Output

The agentic step's output is whatever Claude passes to the `complete` tool. This becomes available as <code v-pre>{{output}}</code> for subsequent steps.

If Claude doesn't call `complete` (reaches max iterations or stops), the last text response is used as fallback.

## Best practices

1. **Set reasonable limits** - Use `max_iterations` to control costs
2. **Provide context** - Include environment info, available tools, constraints
3. **Start with `auto_execute: false`** - Review commands until you trust the workflow
4. **Capture prereq output** - Run tests/checks first and pass results to the agent
5. **Be specific** - Clear instructions lead to better results
