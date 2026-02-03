# Your First Command

In this tutorial, you'll create a simple project command that builds and runs a Go application. No AI required.

## What we're building

A `run` command that:
1. Builds the project
2. Runs the resulting binary

## Step 1: Create the config file

Create a file called `xcommands.yaml` in your project folder:

```yaml
run:
  description: Build and run the application
  steps:
    - exec:
        command: go build -o app . && ./app
```

That's it. Now run:

```bash
x run
```

## Step 2: Add arguments

Let's pass arguments to the binary:

```yaml
run:
  description: Build and run the application
  args:
    - name: flags
      description: Flags to pass to the application
      rest: true
  steps:
    - exec:
        command: go build -o app . && ./app {{args.flags}}
```

Now you can run:

```bash
x run --verbose --port 8080
```

The `rest: true` option captures all remaining arguments as a single string.

## Step 3: Add a confirmation prompt

For commands that do something destructive, add a confirmation:

```yaml
clean:
  description: Remove build artifacts
  steps:
    - exec:
        command: rm -rf ./bin ./dist
        confirm: true
```

Now `x clean` will show the command and ask before running.

## Step 4: Handle different operating systems

The commands above won't work on Windows. Let's fix that:

```yaml
run:
  description: Build and run the application
  args:
    - name: flags
      description: Flags to pass to the application
      rest: true
  steps:
    - exec:
        command: go build -o app . && ./app {{args.flags}}
        windows: go build -o app.exe . && app.exe {{args.flags}}

clean:
  description: Remove build artifacts
  steps:
    - exec:
        command: rm -rf ./bin ./dist
        windows: rmdir /s /q bin dist
        confirm: true
```

You can also specify `darwin` (macOS) and `linux` separately if needed.

## Step 5: Chain multiple commands

Split into separate steps when you need different options:

```yaml
build-and-test:
  description: Build the project then run tests
  steps:
    - exec:
        command: go build -o app .
    - exec:
        command: go test ./...
```

## Complete example

Here's a complete `xcommands.yaml` for a Go project:

```yaml
build:
  description: Build the binary
  steps:
    - exec:
        command: go build -o app .
        windows: go build -o app.exe .

run:
  description: Build and run
  args:
    - name: flags
      rest: true
  steps:
    - exec:
        command: go build -o app . && ./app {{args.flags}}
        windows: go build -o app.exe . && app.exe {{args.flags}}

test:
  description: Run tests
  steps:
    - exec:
        command: go test ./...

test-coverage:
  description: Run tests with coverage
  steps:
    - exec:
        command: go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

clean:
  description: Remove build artifacts
  steps:
    - exec:
        command: rm -f app coverage.out
        windows: del /f app.exe coverage.out
        confirm: true
```

## Next steps

- [Building a Python Fixer](/guide/python-fixer) - Add AI to your commands
- [exec reference](/reference/exec-steps) - All exec options
- [Variables](/reference/variables) - Available template variables
