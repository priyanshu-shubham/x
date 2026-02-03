# Examples

A collection of useful commands you can add to your config.

## Git

### Generate commit message

```yaml
commit:
  description: Generate a commit message from staged changes
  steps:
    - exec:
        command: git diff --cached
        silent: true
    - llm:
        system: |
          Generate a concise git commit message for these changes.
          Use conventional commits format (feat:, fix:, docs:, chore:, etc).
          Output only the commit message, nothing else.
        prompt: "{{output}}"
```

### Summarize recent commits

```yaml
changelog:
  description: Summarize recent commits
  args:
    - name: count
      description: Number of commits (default 10)
  steps:
    - exec:
        command: git log --oneline -{{args.count}} 2>/dev/null || git log --oneline -10
        silent: true
    - llm:
        system: Summarize these commits into a changelog. Group by type (features, fixes, etc).
        prompt: "{{output}}"
```

## Code

### Review a file

```yaml
review:
  description: Review code for issues
  args:
    - name: file
      description: File to review
  steps:
    - exec:
        command: cat {{args.file}}
        windows: type {{args.file}}
        silent: true
    - llm:
        system: |
          Review this code for:
          - Bugs and potential issues
          - Security vulnerabilities
          - Performance problems
          - Code style improvements

          Be specific and actionable.
        prompt: "{{output}}"
```

### Explain code

```yaml
explain:
  description: Explain what code does
  args:
    - name: file
      description: File to explain
  steps:
    - exec:
        command: cat {{args.file}}
        windows: type {{args.file}}
        silent: true
    - llm:
        system: |
          Explain what this code does in plain English.
          Focus on the purpose and high-level logic, not line-by-line details.
          Mention any important patterns or techniques used.
        prompt: "{{output}}"
```

### Add tests

```yaml
add-tests:
  description: Generate tests for a file
  args:
    - name: file
      description: File to generate tests for
  steps:
    - exec:
        command: cat {{args.file}}
        windows: type {{args.file}}
        silent: true
    - llm:
        system: |
          Generate comprehensive unit tests for this code.
          Use the appropriate testing framework for the language.
          Cover edge cases and error conditions.
          Output only the test code.
        prompt: "{{output}}"
```

## Quick utilities

### Ask a question

```yaml
q:
  description: Ask a quick question
  args:
    - name: question
      rest: true
  steps:
    - llm:
        system: Answer concisely. Use code examples when helpful.
        prompt: "{{args.question}}"
```

```bash
x q how do I reverse a string in python
```

### Show my IP

```yaml
ip:
  description: Show public IP address
  steps:
    - exec:
        command: curl -s ifconfig.me
```

### Show listening ports

```yaml
ports:
  description: Show listening ports
  steps:
    - exec:
        command: lsof -i -P -n | grep LISTEN
        linux: ss -tlnp
        windows: netstat -an | findstr LISTENING
```

## Project management

### Build and test

```yaml
check:
  description: Build and run tests
  steps:
    - exec:
        command: go build ./...
    - exec:
        command: go test ./...
```

### Release

```yaml
release:
  description: Tag and push a release
  args:
    - name: version
      description: Version tag (e.g., v1.0.0)
  steps:
    - exec:
        command: git tag {{args.version}} && git push origin main && git push origin {{args.version}}
        confirm: true
```

### Clean build artifacts

```yaml
clean:
  description: Remove build artifacts
  steps:
    - exec:
        command: rm -rf ./bin ./dist ./coverage.out
        windows: rmdir /s /q bin dist 2>nul & del /f coverage.out 2>nul
        confirm: true
```

## AI agents

### Fix issues

```yaml
fix:
  description: Fix an issue in the codebase
  args:
    - name: issue
      description: What to fix
      rest: true
  steps:
    - agentic:
        system: |
          You are a senior developer. Investigate and fix the issue.
          - Read files to understand the code
          - Make minimal, targeted changes
          - Run tests to verify fixes
          Directory: {{directory}}
          OS: {{os}}
        prompt: "{{args.issue}}"
        max_iterations: 15
        auto_execute: false
```

### Refactor

```yaml
refactor:
  description: Refactor code with AI guidance
  args:
    - name: file
      description: File to refactor
    - name: goal
      description: What to improve
      rest: true
  steps:
    - exec:
        command: cat {{args.file}}
        silent: true
    - agentic:
        system: |
          You are refactoring code. The user wants to: {{args.goal}}

          Current code:
          {{output}}

          Make the changes carefully. Test if possible.
        prompt: "Refactor as described above."
        max_iterations: 10
```

## Documentation

### Summarize a file

```yaml
summarize:
  description: Summarize any file
  args:
    - name: file
      description: File to summarize
  steps:
    - exec:
        command: cat {{args.file}}
        windows: type {{args.file}}
        silent: true
    - llm:
        system: Summarize this content. Be concise but capture the key points.
        prompt: "{{output}}"
```

### Generate README

```yaml
readme:
  description: Generate a README for the project
  steps:
    - exec:
        command: find . -type f -name "*.go" -o -name "*.py" -o -name "*.js" -o -name "*.ts" | head -20 | xargs cat 2>/dev/null | head -200
        silent: true
    - llm:
        system: |
          Generate a README.md for this project based on the code.
          Include:
          - Project name and description
          - Installation instructions
          - Basic usage
          - Key features
        prompt: "{{output}}"
```

## Combining commands

### Review and fix

```yaml
review-and-fix:
  description: Review code then fix issues
  args:
    - name: file
      description: File to review and fix
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
