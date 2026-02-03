# Building a Python Fixer

In this tutorial, you'll build an AI agent that can automatically fix issues in Python code. By the end, you'll be able to run:

```bash
x fix-python "the tests are failing"
```

And Claude will investigate the issue, make changes, and verify the fix.

::: warning API Key Required
This tutorial uses AI features. Make sure you've run `x configure` to set up your API key.
:::

## What we're building

An autonomous agent that:
1. Reads error messages and code
2. Investigates the codebase
3. Makes changes to fix the issue
4. Runs tests to verify the fix

## Step 1: Start simple with LLM

Before building an agent, let's start with a simple LLM call:

```yaml
explain-error:
  description: Explain a Python error
  args:
    - name: error
      description: The error message
      rest: true
  steps:
    - llm:
        system: You are a Python expert. Explain this error and suggest fixes.
        prompt: "{{args.error}}"
```

Try it:

```bash
x explain-error "TypeError: 'NoneType' object is not subscriptable"
```

## Step 2: Add context from files

Let's read the actual code first:

```yaml
explain-file:
  description: Explain errors in a Python file
  args:
    - name: file
      description: Path to the Python file
  steps:
    - exec:
        command: cat {{args.file}}
        windows: type {{args.file}}
        silent: true
    - llm:
        system: |
          You are a Python expert. Review this code for bugs and issues.
          Explain any problems you find and suggest fixes.
        prompt: "{{output}}"
```

The `silent: true` captures the file content without printing it, then passes it to the LLM.

## Step 3: Build the autonomous agent

Now let's give Claude the ability to run commands:

```yaml
fix-python:
  description: Fix issues in Python code
  args:
    - name: issue
      description: What to fix
      rest: true
  steps:
    - agentic:
        system: |
          You are a senior Python developer. Your job is to fix the reported issue.

          You have access to a shell. Use it to:
          1. Read files to understand the codebase
          2. Run tests to see what's failing
          3. Make changes to fix the issue
          4. Run tests again to verify the fix

          Environment:
          - OS: {{os}}
          - Directory: {{directory}}
          - Shell: {{shell}}

          Tips:
          - Use `cat` to read files
          - Use `pytest` or `python -m pytest` to run tests
          - Use `sed` or a heredoc to edit files
          - Be careful with indentation in Python
        prompt: "{{args.issue}}"
        max_iterations: 15
        auto_execute: false
```

Try it:

```bash
x fix-python "the login function returns None instead of the user object"
```

Claude will:
1. Ask to read relevant files
2. Ask to run tests to see the failure
3. Ask to make changes
4. Ask to run tests again

Each command requires your approval (because `auto_execute: false`).

## Step 4: Add test output automatically

Let's run tests first and pass the output to the agent:

```yaml
fix-python:
  description: Fix failing Python tests
  args:
    - name: context
      description: Additional context about the issue
      rest: true
  steps:
    - id: tests
      exec:
        command: python -m pytest --tb=short 2>&1 || true
        silent: true
    - agentic:
        system: |
          You are a senior Python developer. Fix the failing tests.

          Test output:
          {{steps.tests.output}}

          Additional context from user:
          {{args.context}}

          Environment: {{os}}, {{shell}}, {{directory}}
        prompt: "Fix the failing tests shown above."
        max_iterations: 15
        auto_execute: false
```

Now `x fix-python` automatically runs tests first and passes the output to the agent.

## Step 5: Make it safer with auto-execute for reads

You might trust Claude to read files but not write them:

```yaml
fix-python:
  description: Fix failing Python tests
  steps:
    - id: tests
      exec:
        command: python -m pytest --tb=short 2>&1 || true
        silent: true
    - agentic:
        system: |
          You are a senior Python developer. Fix the failing tests.

          Test output:
          {{steps.tests.output}}

          IMPORTANT: You can read files freely, but ALWAYS explain what
          changes you want to make before running any write commands.
          The user will approve or deny each write operation.

          Environment: {{os}}, {{directory}}
        prompt: "Fix the failing tests."
        max_iterations: 20
        auto_execute: false
```

## Complete example

Here's the final, polished version:

```yaml
fix-python:
  description: Fix issues in Python code
  args:
    - name: issue
      description: What to fix (optional, runs tests if not provided)
      rest: true
  steps:
    - id: tests
      exec:
        command: python -m pytest --tb=short 2>&1 | head -100 || true
        silent: true
    - agentic:
        system: |
          You are a senior Python developer. Your task is to fix code issues.

          Current test output:
          ```
          {{steps.tests.output}}
          ```

          User's description of the issue:
          {{args.issue}}

          Instructions:
          1. If tests are failing, focus on fixing them
          2. Read relevant files to understand the code
          3. Make minimal, targeted changes
          4. Run tests after each change to verify
          5. Stop when all tests pass or the issue is resolved

          Environment: {{os}} | {{directory}}
        prompt: "Fix the issue described above. If no specific issue is mentioned, fix the failing tests."
        max_iterations: 20
        auto_execute: false

lint-fix:
  description: Auto-fix linting issues
  steps:
    - exec:
        command: python -m black . && python -m isort .
    - exec:
        command: python -m flake8 --max-line-length=100 .
```

## Tips for better agents

1. **Limit iterations** - Set a reasonable `max_iterations` to prevent runaway costs
2. **Provide context** - The more context in the system prompt, the better the results
3. **Use `auto_execute: false`** - Review commands before they run, especially for writes
4. **Capture output silently** - Use `silent: true` for intermediate steps
5. **Include environment info** - Tell the agent about OS, directory, available tools

## Next steps

- [agentic reference](/reference/agentic-steps) - All agentic options
- [Variables](/reference/variables) - All available template variables
- [Examples](/examples/) - More example commands
