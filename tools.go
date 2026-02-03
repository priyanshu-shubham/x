package main

import (
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

// GetShellTool returns the shell command execution tool definition
func GetShellTool() anthropic.ToolUnionParam {
	tv := GetTemplateValues()

	tool := anthropic.ToolUnionParamOfTool(
		anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "The shell command to execute",
				},
			},
			Required: []string{"command"},
		},
		ToolShell,
	)

	description := fmt.Sprintf(`Execute a shell command and return the output.

Environment:
- OS: %s
- Architecture: %s
- Shell: %s
- Current directory: %s
- User: %s

What the user sees:
- The command is shown to the user (either for confirmation or as "Executing: <cmd>")
- Command output is NOT shown to the user, only returned to you
- Your text responses ARE shown to the user (rendered as markdown)

Since command output is hidden from the user, summarize important results in your responses.`, tv.OS, tv.Arch, tv.Shell, tv.Directory, tv.User)

	tool.OfTool.Description = anthropic.String(description)
	return tool
}

// GetCompleteTool returns the completion signal tool definition
func GetCompleteTool() anthropic.ToolUnionParam {
	tool := anthropic.ToolUnionParamOfTool(
		anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"output": map[string]any{
					"type":        "string",
					"description": "The final output or result of the task",
				},
			},
			Required: []string{"output"},
		},
		ToolComplete,
	)
	tool.OfTool.Description = anthropic.String(`Signal that the workflow is complete. Call this when you have finished the task.

The output you provide here becomes the result of this step and may be passed to subsequent steps in the pipeline. Provide the actual output/result, not a description of what you did.

Communicate progress and explanations in your text responses BEFORE calling this.`)
	return tool
}

// BuildAgenticTools returns the tool list for agentic steps
func BuildAgenticTools() []anthropic.ToolUnionParam {
	return []anthropic.ToolUnionParam{
		GetShellTool(),
		GetCompleteTool(),
	}
}
