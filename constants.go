package main

// Version (set via ldflags during build)
var Version = "dev"

// Model constants
const (
	DefaultModelAPI    = "claude-sonnet-4-5-20250929"
	DefaultModelVertex = "claude-sonnet-4-5@20250929"
	DefaultMaxTokens   = 1024
)

// AuthType represents the authentication method
type AuthType string

const (
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeVertex AuthType = "vertex"
)

// Config file names
const (
	ConfigFileName          = "config.json"
	CommandsFileName        = "commands.yaml"
	LocalCommandsFileName   = "xcommands.yaml"
	AppConfigDir            = "x"
)

// Command names (reserved - cannot be used as custom command names)
const (
	CmdConfigure = "configure"
	CmdCommands  = "commands"
	CmdUsage     = "usage"
	CmdUpgrade   = "upgrade"
	CmdVersion   = "version"
)

// GitHub repository
const (
	GitHubOwner = "priyanshu-shubham"
	GitHubRepo  = "x"
)

// Default shells by OS
const (
	DefaultShellUnix    = "/bin/sh"
	DefaultShellWindows = "cmd.exe"
)

// OS identifiers
const (
	OSWindows = "windows"
	OSDarwin  = "darwin"
)

// Template placeholders
const (
	PlaceholderTime      = "{{time}}"
	PlaceholderDate      = "{{date}}"
	PlaceholderDateTime  = "{{datetime}}"
	PlaceholderDirectory = "{{directory}}"
	PlaceholderOS        = "{{os}}"
	PlaceholderArch      = "{{arch}}"
	PlaceholderShell     = "{{shell}}"
	PlaceholderUser      = "{{user}}"
)

// Time formats
const (
	TimeFormat     = "15:04:05"
	DateFormat     = "2006-01-02"
	DateTimeFormat = "2006-01-02 15:04:05"
)

// File permissions
const (
	ConfigFilePerms = 0600
	DirPerms        = 0755
	CommandsPerms   = 0644
)

// Pipeline/Agentic constants
const (
	DefaultMaxIterations = 10
	AgenticMaxTokens     = 4096
	ToolShell            = "shell"
	ToolComplete         = "complete"
)
