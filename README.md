<p align="center">
  <img src="docs/public/logo.svg" width="80" height="80" alt="x logo">
</p>

<h1 align="center">x</h1>

<p align="center">
  Plain English to shell commands. Build custom AI workflows.
  <br>
  <a href="https://priyanshu-shubham.github.io/x/"><strong>Documentation</strong></a>
</p>

```
$ x find files modified in the last day

Summary: Finds all regular files modified within the last 24 hours

┃ find . -mtime -1 -type f

Run this command? [Y/n]:
```

`x` does three things:

1. **Instant shell commands** - Describe what you want, get the right command for your OS
2. **Custom workflows** - Build your own commands that chain shell scripts, LLM calls, and autonomous agents
3. **Project command runner** - Define shortcuts in `xcommands.yaml` and run them with `x build`, `x test`, etc. Works without AI too - like npm scripts or Makefiles, but simpler

## Install

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/priyanshu-shubham/x/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/priyanshu-shubham/x/main/install.ps1 | iex
```

**From source:**
```bash
go install github.com/priyanshu-shubham/x@latest
```

**Optional:** Configure an API key if you want to use AI features:
```
x configure
```

You can use an Anthropic API key or Google Cloud Vertex AI. Skip this if you only need the command runner.

## Basic usage

Just type `x` followed by what you want to do:

```bash
x list all png files larger than 1mb
x find processes using port 3000
x show disk usage sorted by size
x compress all images in this folder
```

You'll see the command and can choose whether to run it.

---

## Tutorial: Build your first custom command

Let's create a command that summarizes any file. By the end, you'll be able to run:

```
x summarize README.md
```

### Step 1: Create the config file

Create a file called `xcommands.yaml` in your project folder:

```yaml
summarize:
  description: Summarize any file
  args:
    - name: file
      description: Path to the file
  steps:
    - exec:
        command: cat {{args.file}}
        silent: true
    - llm:
        system: Summarize the following content. Be concise.
        prompt: "{{output}}"
```

That's it. Now `x summarize myfile.txt` will read the file and send it to Claude for summarization.

### What's happening here?

1. **args** defines what arguments your command takes. `{{args.file}}` gets replaced with whatever the user passes in.

2. **steps** run in order. Each step can be:
   - `exec` - run a shell command
   - `llm` - call Claude
   - `agentic` - let Claude run commands autonomously (more on this later)

3. **silent: true** means don't print the output, just pass it to the next step.

4. **{{output}}** contains whatever the previous step produced.

### Step 2: Make it cross-platform

The `cat` command doesn't work on Windows. Let's fix that:

```yaml
summarize:
  description: Summarize any file
  args:
    - name: file
      description: Path to the file
  steps:
    - exec:
        command: cat {{args.file}}
        windows: type {{args.file}}
        silent: true
    - llm:
        system: Summarize the following content. Be concise.
        prompt: "{{output}}"
```

Now it uses `type` on Windows and `cat` everywhere else. You can also specify `darwin` (macOS) and `linux` separately if needed.

### Step 3: Add a confirmation step

Want to review a command before it runs? Add `confirm: true`:

```yaml
delete-old:
  description: Delete files older than 30 days
  steps:
    - llm:
        system: Generate a command to delete files older than 30 days. Output only the command.
        prompt: "Delete old files in {{directory}}"
        silent: true
    - exec:
        command: "{{output}}"
        confirm: true
```

This generates a delete command with AI, shows it to you, and only runs it if you approve.

---

## Four types of steps

### 1. `exec` - Run shell commands

```yaml
steps:
  - exec:
      command: echo hello
```

**Options:**
- `command` - the command to run (required)
- `windows`, `darwin`, `linux` - OS-specific overrides
- `silent: true` - capture output without printing it
- `confirm: true` - ask before running
- `summary` - description shown before confirm (supports interpolation)
- `risk` - risk level: `none`, `low`, `medium`, `high`
- `safer` - safer alternative shown for risky commands

When `silent` is false (the default), output streams to your terminal in real-time.

**Smart confirmation:** When `confirm: true` with `risk` set, the prompt changes based on risk level:
- `none`/`low`: Defaults to Yes (`[Y/n]`)
- `medium`/`high`: Defaults to No (`[y/N]`), requires explicit confirmation

### 2. `llm` - Call Claude

```yaml
steps:
  - llm:
      system: You are a helpful assistant.
      prompt: "{{args.question}}"
```

**Options:**
- `system` - the system prompt
- `prompt` - the user prompt
- `silent: true` - capture response without printing it

### 3. `agentic` - Autonomous mode

This is the powerful one. Claude gets access to a shell and can run commands on its own to complete complex tasks:

```yaml
fix:
  description: Fix issues in the codebase
  args:
    - name: issue
      description: What to fix
      rest: true
  steps:
    - agentic:
        system: |
          You are a senior developer. Use shell commands to explore
          the codebase and fix the issue. You have access to: {{os}}, {{shell}}
          Current directory: {{directory}}
        prompt: "{{args.issue}}"
        max_iterations: 10
        auto_execute: false
```

```
$ x fix the login button is broken
```

Claude will read files, understand the code, make changes, and verify they work - asking for permission before each command (unless you set `auto_execute: true`).

**Options:**
- `system`, `prompt` - same as llm
- `max_iterations` - how many commands Claude can run (default: 10)
- `auto_execute: true` - run commands without asking (be careful!)

### 4. `subcommand` - Reuse other commands

Call another command as a step. Great for composing complex workflows from simpler building blocks:

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

**Options:**
- `name` - the command to call
- `args` - list of arguments (supports variable interpolation)
- `silent: true` - don't print "Running subcommand:" message

---

## Passing data between steps

### Using `{{output}}`

The simplest way - each step can access the previous step's output:

```yaml
steps:
  - exec:
      command: git diff --cached
      silent: true
  - llm:
      system: Generate a commit message for these changes.
      prompt: "{{output}}"
```

### Using named steps

When you need output from a specific step (not just the previous one), give it an `id`:

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
      system: Compare these two files.
      prompt: |
        README:
        {{steps.readme.output}}

        CHANGELOG:
        {{steps.changelog.output}}
```

---

## Available variables

Use these anywhere in your prompts:

| Variable | What it contains |
|----------|------------------|
| `{{args.name}}` | Argument passed by the user |
| `{{output}}` | Raw output from the previous step |
| `{{output.field}}` | JSON field from previous step (if output is JSON) |
| `{{steps.id.output}}` | Output from a specific named step |
| `{{steps.id.field}}` | JSON field from a named step |
| `{{directory}}` | Current working directory |
| `{{os}}` | Operating system (darwin, linux, windows) |
| `{{arch}}` | CPU architecture |
| `{{shell}}` | User's shell |
| `{{user}}` | Username |
| `{{date}}` | Current date |
| `{{time}}` | Current time |
| `{{datetime}}` | Current date and time |

---

## Arguments

Define what your command accepts:

```yaml
mycommand:
  args:
    - name: file
      description: The file to process
    - name: query
      description: What to search for
      rest: true
```

`rest: true` captures everything remaining as a single string. Useful for natural language input:

```
x mycommand data.txt find all the TODO comments
```

Here, `file` = `data.txt` and `query` = `find all the TODO comments`.

---

## Config files

### Global config

Your global commands live in:
- **macOS:** `~/Library/Application Support/x/commands.yaml`
- **Linux:** `~/.config/x/commands.yaml`
- **Windows:** `%LOCALAPPDATA%\x\commands.yaml`

Edit with `x commands` to open in your default editor.

### Project config

Create `xcommands.yaml` in any directory. Commands defined here are available when you're in that directory (or any subdirectory).

**How merging works:**

All configs are merged together. If you're in `/projects/myapp/src/`, x loads:
1. Global `commands.yaml`
2. `/projects/xcommands.yaml` (if it exists)
3. `/projects/myapp/xcommands.yaml` (if it exists)

Later files override earlier ones.

**Built-in commands:**

The `shell` and `new` commands are built into the binary and always up-to-date. You don't need to define them - they're automatically available and updated when you upgrade `x`.

Run `x --help` to see where each command comes from:

```
Commands (built-in):
  shell  Generate and run shell commands from natural language (default)
  new    Create a new command with AI assistance

Commands (global):
  mycommand  A custom command you added

Commands (myapp):
  build  Build the project
  test   Run tests
```

You can override built-in commands in your config if you want custom behavior.

---

## Create commands with AI

Don't want to write YAML? Let Claude do it:

```
x new "a command that reviews my git diff and suggests improvements"
```

Claude will generate the YAML and add it to your config file.

---

## Debugging

### See what's happening

```bash
DEBUG=1 x summarize README.md
```

Shows all the prompts being sent to Claude, step execution details, and more.

### Dry run

```bash
DRYRUN=1 x "delete all tmp files"
```

Shows what would happen without actually running anything. Good for testing dangerous commands.

### Combine them

```bash
DEBUG=1 DRYRUN=1 x fix "update the readme"
```

---

## Example commands

Here are some useful commands you can add to your config:

### Git commit message generator

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
          Use conventional commits (feat:, fix:, docs:, etc).
          Output only the message.
        prompt: "{{output}}"
```

### Code reviewer

```yaml
review:
  description: Review a file for issues
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
          Review this code for bugs, security issues, and improvements.
          Be specific and actionable.
        prompt: "{{output}}"
```

### Explain code

```yaml
explain:
  description: Explain what a file does
  args:
    - name: file
      description: File to explain
  steps:
    - exec:
        command: cat {{args.file}}
        windows: type {{args.file}}
        silent: true
    - llm:
        system: Explain what this code does in plain English. Focus on the purpose, not line-by-line details.
        prompt: "{{output}}"
```

### Quick question

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

```
x q how do I reverse a string in python
```

### Autonomous bug fixer

```yaml
fix:
  description: Fix an issue in the codebase
  args:
    - name: issue
      rest: true
  steps:
    - agentic:
        system: |
          You are a senior developer. Investigate and fix the issue.
          Read files, understand the code, make changes, and verify they work.
          Directory: {{directory}}
          OS: {{os}}
        prompt: "{{args.issue}}"
        max_iterations: 15
        auto_execute: false
```

---

## Built-in commands

| Command | What it does |
|---------|--------------|
| `x configure` | Set up API credentials |
| `x commands` | Edit global commands in your editor |
| `x usage` | Show token usage and estimated cost |
| `x upgrade` | Upgrade to the latest version |
| `x version` | Show current version |
| `x --help` | List all available commands |
| `x <cmd> --help` | Show help for a specific command |

---

## Tips

- **Start simple.** Get a basic command working, then add complexity.
- **Use `silent: true`** for intermediate steps whose output you don't need to see.
- **Use `confirm: true`** for any command that modifies or deletes things.
- **Use `agentic` sparingly.** It's powerful but uses more tokens.
- **Commit your `xcommands.yaml`** so your team can use the same commands.
- **Messed up your config?** Delete it and run any command - it'll be recreated with defaults.

---

## Troubleshooting

**Command not found after install**

Make sure `~/.local/bin` (Linux/macOS) is in your PATH.

**API errors**

Run `x configure` to check your credentials. Make sure you have credits/quota remaining.

**Config syntax errors**

YAML is picky about indentation. Use 2 spaces, not tabs. Run with `DEBUG=1` to see parsing errors.

**Command hangs**

If running a long-lived process (like a server) as a non-final step, it will block. Long-running processes should be the last step, or run them in the background with `&`.
