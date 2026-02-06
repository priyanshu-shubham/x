package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

// isDebug returns true if DEBUG environment variable is set
func isDebug() bool {
	val := os.Getenv("DEBUG")
	return val != "" && val != "0" && val != "false"
}

// isDryRun returns true if DRYRUN environment variable is set
func isDryRun() bool {
	val := os.Getenv("DRYRUN")
	return val != "" && val != "0" && val != "false"
}

// debugLog prints debug information when DEBUG is enabled
func debugLog(format string, args ...any) {
	if isDebug() {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// debugSection prints a debug section header
func debugSection(title string) {
	if isDebug() {
		fmt.Printf("\n[DEBUG] === %s ===\n", title)
	}
}

// debugPrompt prints prompt content in debug mode
func debugPrompt(label, content string) {
	if isDebug() {
		fmt.Printf("[DEBUG] %s:\n", label)
		// Indent the content
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			fmt.Printf("[DEBUG]   %s\n", line)
		}
	}
}

// isRiskyCommand returns true if the risk level is medium or high
func isRiskyCommand(risk string) bool {
	riskLower := strings.ToLower(risk)
	return strings.HasPrefix(riskLower, "medium") || strings.HasPrefix(riskLower, "high")
}

// printConfirmInfo prints summary, risk, and safer alternative before confirmation
func printConfirmInfo(summary, risk, safer string) {
	// Print summary if present
	if summary != "" {
		fmt.Printf("\n\033[1mSummary:\033[0m %s\n", summary)
	}

	// Only show risk and safer alternative for medium/high risk commands
	if isRiskyCommand(risk) {
		// Print risk with color coding
		riskLower := strings.ToLower(risk)
		var color string
		switch {
		case strings.HasPrefix(riskLower, "medium"):
			color = "\033[38;5;208m" // Orange
		case strings.HasPrefix(riskLower, "high"):
			color = "\033[31m" // Red
		default:
			color = "\033[0m"
		}
		fmt.Printf("\033[1mRisk:\033[0m %s%s\033[0m\n", color, risk)

		// Print safer alternative if present
		if safer != "" {
			fmt.Printf("\033[1mSafer alternative:\033[0m %s\n", safer)
		}
	}
}

// PipelineContext tracks state during pipeline execution
type PipelineContext struct {
	Args        map[string]string // Parsed argument name -> value
	StepOutputs map[string]string // Step ID -> output
	LastOutput  string            // Output from previous step
}

// NewPipelineContext creates a new pipeline context
func NewPipelineContext() *PipelineContext {
	return &PipelineContext{
		Args:        make(map[string]string),
		StepOutputs: make(map[string]string),
	}
}

// RunPipeline executes all steps in a command pipeline
// Returns the final step's output and any error
// If captureOutput is true, the last step will use streaming instead of interactive mode
func RunPipeline(client anthropic.Client, authType AuthType, config *CommandsConfig, cmd Command, userArgs []string, captureOutput bool) (string, error) {
	ctx := NewPipelineContext()

	if isDryRun() {
		fmt.Println("[DRYRUN] Dry run mode - no commands will be executed")
	}

	debugLog("Starting pipeline with %d steps", len(cmd.Steps))
	debugLog("User args: %v", userArgs)

	// Parse arguments
	if err := parseArgs(ctx, cmd.Args, userArgs); err != nil {
		return "", err
	}

	debugLog("Parsed args: %v", ctx.Args)

	// Execute each step
	for i, step := range cmd.Steps {
		var output string
		var err error

		stepID := step.ID
		if stepID == "" {
			stepID = fmt.Sprintf("step-%d", i+1)
		}

		isLastStep := i == len(cmd.Steps)-1

		switch {
		case step.Exec != nil:
			debugSection(fmt.Sprintf("Step %d: exec (id=%s)", i+1, stepID))
			output, err = runExecStep(ctx, step.Exec, isLastStep, captureOutput)
		case step.LLM != nil:
			debugSection(fmt.Sprintf("Step %d: llm (id=%s)", i+1, stepID))
			output, err = runLLMStep(client, authType, ctx, step.LLM)
		case step.Agentic != nil:
			debugSection(fmt.Sprintf("Step %d: agentic (id=%s)", i+1, stepID))
			output, err = runAgenticStep(client, authType, ctx, step.Agentic)
		case step.Subcommand != nil:
			debugSection(fmt.Sprintf("Step %d: subcommand (id=%s)", i+1, stepID))
			output, err = runSubcommandStep(client, authType, config, ctx, step.Subcommand)
		default:
			return "", fmt.Errorf("step %d has no valid type (exec, llm, agentic, or subcommand)", i+1)
		}

		if err != nil {
			return "", fmt.Errorf("step %d failed: %w", i+1, err)
		}

		// Store output
		ctx.LastOutput = output
		if step.ID != "" {
			ctx.StepOutputs[step.ID] = output
		}

		debugLog("Step output length: %d bytes", len(output))
	}

	if isDryRun() {
		fmt.Println("[DRYRUN] Dry run complete")
	}

	return ctx.LastOutput, nil
}

// parseArgs parses user arguments into the pipeline context
func parseArgs(ctx *PipelineContext, argDefs []Arg, userArgs []string) error {
	argIndex := 0

	for _, argDef := range argDefs {
		if argIndex >= len(userArgs) {
			return fmt.Errorf("missing required argument: %s", argDef.Name)
		}

		if argDef.Rest {
			// Capture all remaining args as one string
			ctx.Args[argDef.Name] = strings.Join(userArgs[argIndex:], " ")
			argIndex = len(userArgs)
		} else {
			ctx.Args[argDef.Name] = userArgs[argIndex]
			argIndex++
		}
	}

	return nil
}

// getOSCommand returns the appropriate command for the current OS
func getOSCommand(step *ExecStep) string {
	tv := GetTemplateValues()

	switch tv.OS {
	case "windows":
		if step.Windows != "" {
			return step.Windows
		}
	case "darwin":
		if step.Darwin != "" {
			return step.Darwin
		}
	case "linux":
		if step.Linux != "" {
			return step.Linux
		}
	}

	return step.Command
}

// runExecStep executes a shell command step
// If isLastStep is true and captureOutput is false, runs interactively with terminal connected
// If captureOutput is true, always captures output (for command chaining)
func runExecStep(ctx *PipelineContext, step *ExecStep, isLastStep bool, captureOutput bool) (string, error) {
	command := getOSCommand(step)
	var err error
	command, err = interpolateVariables(command, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to interpolate command: %w", err)
	}

	debugLog("Command: %s", command)
	debugLog("Confirm: %v, Silent: %v, IsLastStep: %v, CaptureOutput: %v", step.Confirm, step.Silent, isLastStep, captureOutput)

	if isDryRun() {
		fmt.Printf("[DRYRUN] Would execute: %s\n", command)
		// Show optional fields if present
		if step.Summary != "" {
			summary, _ := interpolateVariables(step.Summary, ctx)
			if summary != "" {
				fmt.Printf("[DRYRUN] Summary: %s\n", summary)
			}
		}
		if step.Risk != "" {
			risk, _ := interpolateVariables(step.Risk, ctx)
			if risk != "" {
				fmt.Printf("[DRYRUN] Risk: %s\n", risk)
			}
		}
		return "[dry run - no output]", nil
	}

	if step.Confirm {
		// Interpolate optional summary/risk/safer fields
		summary, _ := interpolateVariables(step.Summary, ctx)
		risk, _ := interpolateVariables(step.Risk, ctx)
		safer, _ := interpolateVariables(step.Safer, ctx)

		// Display confirmation info
		printConfirmInfo(summary, risk, safer)
		printCommandForConfirm(command)

		reader := bufio.NewReader(os.Stdin)
		isHighRisk := isRiskyCommand(risk)

		if isHighRisk {
			// For medium/high risk: default to No, require explicit Y
			fmt.Print("Run this command? [y/N]: ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				os.Exit(0)
			}
		} else {
			// For none/low risk: default to Yes
			fmt.Print("Run this command? [Y/n]: ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "n" || response == "no" {
				fmt.Println("Cancelled.")
				os.Exit(0)
			}
		}
		fmt.Println()
	}

	// For the last step without silent, run interactively with terminal connected
	// Unless captureOutput is true (nested command call), then use streaming to capture output
	if isLastStep && !step.Silent {
		if !step.Confirm {
			// Show the command being executed (confirm already showed it)
			printExecCommand(command)
		}
		if captureOutput {
			// Need to capture output for chaining, use streaming
			output, err := RunShellCommandStreaming(command)
			return output, err
		}
		// Top-level call, run interactively
		err := RunShellCommand(command)
		return "", err
	}

	// Show the command being executed unless silent or already confirmed
	if !step.Silent && !step.Confirm {
		printExecCommand(command)
	}

	// Execute command: stream output if not silent, otherwise capture silently
	if step.Silent {
		output, err := RunShellCommandWithOutput(command)
		return output, err
	}

	// Stream dimmed output to terminal while capturing
	output, err := RunShellCommandStreaming(command)
	return output, err
}

// runLLMStep executes a single LLM call step
func runLLMStep(client anthropic.Client, authType AuthType, ctx *PipelineContext, step *LLMStep) (string, error) {
	systemPrompt, err := interpolateVariables(step.System, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to interpolate system prompt: %w", err)
	}
	systemPrompt = ApplyTemplate(systemPrompt)

	prompt, err := interpolateVariables(step.Prompt, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to interpolate user prompt: %w", err)
	}
	prompt = ApplyTemplate(prompt)

	debugPrompt("System prompt", systemPrompt)
	debugPrompt("User prompt", prompt)

	if isDryRun() {
		fmt.Println("[DRYRUN] Would call LLM with:")
		fmt.Println("[DRYRUN]   System prompt length:", len(systemPrompt), "bytes")
		fmt.Println("[DRYRUN]   User prompt length:", len(prompt), "bytes")
		return "[dry run - no LLM response]", nil
	}

	response, err := GenerateResponse(client, authType, systemPrompt, prompt)
	if err != nil {
		return "", err
	}

	// Render markdown for LLM output unless silent
	if !step.Silent {
		renderMarkdown(response)
	}

	return response, nil
}

// runAgenticStep executes a multi-turn agentic loop
func runAgenticStep(client anthropic.Client, authType AuthType, ctx *PipelineContext, step *AgenticStep) (string, error) {
	systemPrompt, err := interpolateVariables(step.System, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to interpolate system prompt: %w", err)
	}
	systemPrompt = ApplyTemplate(systemPrompt)

	prompt, err := interpolateVariables(step.Prompt, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to interpolate user prompt: %w", err)
	}
	prompt = ApplyTemplate(prompt)

	maxIterations := step.MaxIterations
	if maxIterations <= 0 {
		maxIterations = DefaultMaxIterations
	}

	debugPrompt("System prompt", systemPrompt)
	debugPrompt("User prompt", prompt)
	debugLog("Max iterations: %d", maxIterations)
	debugLog("Auto execute: %v", step.AutoExecute)

	if isDryRun() {
		fmt.Println("[DRYRUN] Would start agentic loop with:")
		fmt.Println("[DRYRUN]   System prompt length:", len(systemPrompt), "bytes")
		fmt.Println("[DRYRUN]   User prompt length:", len(prompt), "bytes")
		fmt.Println("[DRYRUN]   Max iterations:", maxIterations)
		fmt.Println("[DRYRUN]   Auto execute:", step.AutoExecute)
		fmt.Println("[DRYRUN]   Tools: shell, complete")
		return "[dry run - no agentic execution]", nil
	}

	tools := BuildAgenticTools()
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
	}

	var lastTextBlock string // Track last text for fallback output
	var finalOutput string   // The actual output (from complete tool)
	bgCtx := context.Background()

	for iteration := 0; iteration < maxIterations; iteration++ {
		debugLog("Agentic iteration %d/%d", iteration+1, maxIterations)

		response, err := client.Messages.New(bgCtx, anthropic.MessageNewParams{
			Model:     anthropic.Model(ModelForAuth(authType)),
			MaxTokens: AgenticMaxTokens,
			System: []anthropic.TextBlockParam{
				{Text: systemPrompt},
			},
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return lastTextBlock, err
		}

		// Track usage
		trackUsage(response.Usage)

		// Process response content
		var toolResults []anthropic.ContentBlockParamUnion
		var assistantBlocks []anthropic.ContentBlockParamUnion
		completed := false

		for _, block := range response.Content {
			switch block.Type {
			case "text":
				text := block.Text
				lastTextBlock = text // Keep track of last text for fallback
				renderMarkdown(text)
				assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(text))

			case "tool_use":
				toolName := block.Name
				toolID := block.ID
				input := block.Input

				debugLog("Tool call: %s (id=%s)", toolName, toolID)
				debugLog("Tool input: %s", string(input))

				// Add tool use block to assistant message
				assistantBlocks = append(assistantBlocks, anthropic.NewToolUseBlock(toolID, input, toolName))

				switch toolName {
				case ToolShell:
					result, isError := handleShellTool(input, step.AutoExecute)
					toolResults = append(toolResults, anthropic.NewToolResultBlock(toolID, result, isError))

				case ToolComplete:
					finalOutput = extractOutput(input)
					completed = true
					toolResults = append(toolResults, anthropic.NewToolResultBlock(toolID, "Workflow completed.", false))
				}
			}
		}

		if completed {
			return finalOutput, nil
		}

		// If there were tool calls, add assistant message and tool results
		if len(toolResults) > 0 {
			messages = append(messages, anthropic.NewAssistantMessage(assistantBlocks...))
			messages = append(messages, anthropic.NewUserMessage(toolResults...))
		}

		// If stop reason is end_turn without tool use, we're done
		if response.StopReason == "end_turn" && len(toolResults) == 0 {
			fmt.Printf("\n\033[33m⚠ Agent finished without calling complete tool\033[0m\n")
			return lastTextBlock, nil
		}
	}

	// Max iterations reached without completion
	fmt.Printf("\n\033[33m⚠ Agent reached max iterations (%d) without completing\033[0m\n", maxIterations)
	return lastTextBlock, nil
}

// runSubcommandStep executes another command as a step
func runSubcommandStep(client anthropic.Client, authType AuthType, config *CommandsConfig, ctx *PipelineContext, step *SubcommandStep) (string, error) {
	// Look up the command
	cmd, ok := config.Commands[step.Name]
	if !ok {
		return "", fmt.Errorf("command not found: %s", step.Name)
	}

	// Interpolate args
	var args []string
	for _, arg := range step.Args {
		interpolated, err := interpolateVariables(arg, ctx)
		if err != nil {
			return "", fmt.Errorf("failed to interpolate arg %q: %w", arg, err)
		}
		interpolated = ApplyTemplate(interpolated)
		args = append(args, interpolated)
	}

	debugLog("Calling command: %s with args: %v", step.Name, args)

	if isDryRun() {
		fmt.Printf("[DRYRUN] Would call command: %s %v\n", step.Name, args)
		return "[dry run - no command execution]", nil
	}

	if !step.Silent {
		fmt.Printf("\n\033[1;35m❯ Running command:\033[0m %s %s\n", step.Name, strings.Join(args, " "))
	}

	// Run the command pipeline
	output, err := RunPipeline(client, authType, config, cmd, args, true)
	if err != nil {
		return "", fmt.Errorf("command %s failed: %w", step.Name, err)
	}

	return output, nil
}

// handleShellTool processes a shell tool call
func handleShellTool(input json.RawMessage, autoExecute bool) (string, bool) {
	var params struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return fmt.Sprintf("Error parsing tool input: %v", err), true
	}

	if !autoExecute {
		printCommandForConfirm(params.Command)
		fmt.Print("Run this command? [Y/n]: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "" && response != "y" && response != "yes" {
			return "Command execution cancelled by user.", false
		}
		fmt.Println()
	} else {
		printExecCommand(params.Command)
	}

	output, err := RunShellCommandWithOutput(params.Command)
	if err != nil {
		return fmt.Sprintf("Error: %v\nOutput: %s", err, output), true
	}

	return output, false
}

// extractOutput extracts the output from the complete tool input
func extractOutput(input json.RawMessage) string {
	var params struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return ""
	}
	return params.Output
}

// trackUsage tracks token usage from a response
func trackUsage(usage anthropic.Usage) {
	if usage.InputTokens > 0 || usage.OutputTokens > 0 {
		usageData, err := LoadUsage()
		if err == nil {
			usageData.Add(usage.InputTokens, usage.OutputTokens, 0, usage.CacheCreationInputTokens, usage.CacheReadInputTokens)
			usageData.Save()
		}
	}
}

// interpolateVariables replaces {{args.X}}, {{output}}, and {{steps.X.output}} placeholders
func interpolateVariables(text string, ctx *PipelineContext) (string, error) {
	var interpolateErr error

	// Replace {{args.name}} patterns
	argsPattern := regexp.MustCompile(`\{\{args\.([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	text = argsPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract argument name
		name := match[7 : len(match)-2] // Remove "{{args." and "}}"
		if value, ok := ctx.Args[name]; ok {
			return value
		}
		return match // Keep original if not found
	})

	// Replace {{steps.id.field}} patterns (including {{steps.id.output}})
	stepsPattern := regexp.MustCompile(`\{\{steps\.([a-zA-Z_][a-zA-Z0-9_-]*)\.([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	text = stepsPattern.ReplaceAllStringFunc(text, func(match string) string {
		if interpolateErr != nil {
			return match
		}
		// Extract step ID and field
		parts := strings.Split(match[2:len(match)-2], ".") // Remove "{{" and "}}"
		if len(parts) >= 3 {
			stepID := parts[1]
			field := parts[2]
			output, ok := ctx.StepOutputs[stepID]
			if !ok {
				return match // Keep original if step not found
			}
			if field == "output" {
				return output // Raw output
			}
			// Try to parse as JSON and extract field
			value, err := extractJSONField(output, field)
			if err != nil {
				interpolateErr = fmt.Errorf("cannot access {{steps.%s.%s}}: %w", stepID, field, err)
				return match
			}
			return value
		}
		return match
	})

	if interpolateErr != nil {
		return "", interpolateErr
	}

	// Replace {{output.field}} patterns (JSON field access)
	outputFieldPattern := regexp.MustCompile(`\{\{output\.([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	text = outputFieldPattern.ReplaceAllStringFunc(text, func(match string) string {
		if interpolateErr != nil {
			return match
		}
		field := match[9 : len(match)-2] // Remove "{{output." and "}}"
		value, err := extractJSONField(ctx.LastOutput, field)
		if err != nil {
			interpolateErr = fmt.Errorf("cannot access {{output.%s}}: %w", field, err)
			return match
		}
		return value
	})

	if interpolateErr != nil {
		return "", interpolateErr
	}

	// Replace {{output}} with last output (raw)
	text = strings.ReplaceAll(text, "{{output}}", ctx.LastOutput)

	return text, nil
}

// extractJSONField parses a JSON string and extracts a field value
func extractJSONField(jsonStr, field string) (string, error) {
	// Strip markdown code blocks if present (```json ... ``` or ``` ... ```)
	jsonStr = stripMarkdownCodeBlock(jsonStr)

	var data map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("output is not valid JSON: %w", err)
	}

	value, ok := data[field]
	if !ok {
		return "", fmt.Errorf("field %q not found in JSON", field)
	}

	// Convert value to string
	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return fmt.Sprintf("%v", v), nil
	case bool:
		return fmt.Sprintf("%v", v), nil
	case nil:
		return "", nil
	default:
		// For complex types, marshal back to JSON
		bytes, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
}

// stripMarkdownCodeBlock removes markdown code block syntax from a string
// Handles ```json ... ```, ```... ```, and plain content
func stripMarkdownCodeBlock(s string) string {
	s = strings.TrimSpace(s)

	// Check for code block with language identifier (```json, ```javascript, etc.)
	if strings.HasPrefix(s, "```") {
		// Find the end of the first line (the opening ```)
		firstNewline := strings.Index(s, "\n")
		if firstNewline == -1 {
			return s // No newline, return as-is
		}

		// Find the closing ```
		lastBackticks := strings.LastIndex(s, "```")
		if lastBackticks > firstNewline {
			// Extract content between opening line and closing ```
			return strings.TrimSpace(s[firstNewline+1 : lastBackticks])
		}
	}

	return s
}

// printExecCommand prints the command being executed in a formatted way
func printExecCommand(command string) {
	fmt.Printf("\n\033[1;34m❯ Executing:\033[0m %s\n", command)
}

// printCommandForConfirm renders a command as a code block for confirmation prompts
func printCommandForConfirm(command string) {
	markdown := fmt.Sprintf("```sh\n%s\n```", command)
	renderMarkdown(markdown)
}

// renderMarkdown renders text as markdown to the terminal
func renderMarkdown(text string) {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(getMarkdownStyle()),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		fmt.Println(text)
		return
	}

	rendered, err := renderer.Render(text)
	if err != nil {
		fmt.Println(text)
		return
	}

	rendered = addCodeBlockBorder(rendered)
	fmt.Print(rendered)
}
