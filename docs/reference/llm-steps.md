# llm Steps

Make a single call to Claude.

## Basic usage

```yaml
steps:
  - llm:
      system: You are a helpful assistant.
      prompt: "What is the capital of France?"
```

## Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `system` | string | - | System prompt (instructions for Claude) |
| `prompt` | string | required | User prompt (the question/task) |
| `silent` | boolean | `false` | Capture response without printing |

## System prompt

The `system` field sets Claude's behavior and context:

```yaml
steps:
  - llm:
      system: |
        You are a senior code reviewer.
        Focus on security issues and performance.
        Be concise and actionable.
      prompt: "Review this code: {{output}}"
```

Tips for system prompts:
- Be specific about the task
- Specify the output format you want
- Mention constraints (length, style, etc.)

## Using with exec

A common pattern is to capture command output, then send it to Claude:

```yaml
commit:
  description: Generate a commit message
  steps:
    - exec:
        command: git diff --cached
        silent: true
    - llm:
        system: |
          Generate a concise git commit message for these changes.
          Use conventional commits format (feat:, fix:, docs:, etc).
          Output only the commit message, nothing else.
        prompt: "{{output}}"
```

## Silent mode

When `silent: true`, the response is captured but not printed:

```yaml
steps:
  - llm:
      system: Generate a shell command to list large files.
      prompt: "Find files over 100MB"
      silent: true
  - exec:
      command: "{{output}}"
      confirm: true
```

This is useful when the LLM generates a command that you want to execute.

## Variable interpolation

All [template variables](/reference/variables) work in both `system` and `prompt`:

```yaml
steps:
  - llm:
      system: |
        You are helping a user on {{os}}.
        Current directory: {{directory}}
        Current time: {{datetime}}
      prompt: "{{args.question}}"
```

## Multi-step conversations

Each `llm` step is independent - there's no conversation history between steps. If you need multi-turn conversation, use [agentic mode](/reference/agentic-steps) instead.

For sequential LLM calls, pass context explicitly:

```yaml
steps:
  - id: analysis
    llm:
      system: Analyze this code for bugs.
      prompt: "{{output}}"
      silent: true
  - llm:
      system: Suggest fixes for these bugs.
      prompt: "{{steps.analysis.output}}"
```

## Output as step result

The LLM's response becomes the step's output, available as <code v-pre>{{output}}</code> in the next step or <code v-pre>{{steps.id.output}}</code> if the step has an `id`.
