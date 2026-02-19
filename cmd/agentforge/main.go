package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/agents"
	"github.com/devaloi/agentforge/internal/config"
	"github.com/devaloi/agentforge/internal/memory"
	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/supervisor"
	"github.com/devaloi/agentforge/internal/tools"
)

var (
	configDir    string
	outputDir    string
	providerFlag string
	modelFlag    string
	verbose      bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "agentforge",
		Short: "Multi-agent orchestration framework",
		Long:  "A multi-agent orchestration framework — a supervisor agent decomposes tasks into a DAG, delegates to specialized sub-agents, manages shared memory, and synthesizes results.",
	}

	runCmd := &cobra.Command{
		Use:   "run [task]",
		Short: "Run the supervisor with a task",
		Args:  cobra.ExactArgs(1),
		RunE:  runTask,
	}
	runCmd.Flags().StringVar(&configDir, "config", "./config/", "Configuration directory")
	runCmd.Flags().StringVar(&outputDir, "output-dir", "./output/", "Output directory for generated files")
	runCmd.Flags().StringVar(&providerFlag, "provider", "", "Override default provider")
	runCmd.Flags().StringVar(&modelFlag, "model", "", "Override default model")
	runCmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed execution trace")

	agentsCmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage agents",
	}
	agentsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List configured agents",
		RunE:  listAgents,
	}
	agentsListCmd.Flags().StringVar(&configDir, "config", "./config/", "Configuration directory")
	agentsCmd.AddCommand(agentsListCmd)

	toolsCmd := &cobra.Command{
		Use:   "tools",
		Short: "Manage tools",
	}
	toolsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List registered tools",
		RunE:  listTools,
	}
	toolsListCmd.Flags().StringVar(&configDir, "config", "./config/", "Configuration directory")
	toolsCmd.AddCommand(toolsListCmd)

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
	}
	configValidateCmd := &cobra.Command{
		Use:   "validate [dir]",
		Short: "Validate configuration files",
		Args:  cobra.ExactArgs(1),
		RunE:  validateConfig,
	}
	configCmd.AddCommand(configValidateCmd)

	rootCmd.AddCommand(runCmd, agentsCmd, toolsCmd, configCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runTask(_ *cobra.Command, args []string) error {
	task := args[0]

	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	cfg, err := config.Load(configDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	provRegistry := provider.NewRegistry()
	for name, pc := range cfg.Providers {
		p := createProvider(name, pc)
		if p != nil {
			provRegistry.Register(p)
		}
	}

	defaultProvider := resolveProvider(provRegistry, cfg)
	if defaultProvider == nil {
		logger.Info("no LLM provider configured, using mock provider for demonstration")
		defaultProvider = createDemoProvider(task)
		provRegistry.Register(defaultProvider)
	}

	model := resolveModel(cfg)
	mem := memory.NewStore()

	factory := buildFactory(defaultProvider, mem, logger)

	sv := supervisor.New(defaultProvider, model, factory, 4, logger)

	fmt.Println("🔵 Supervisor: Planning task decomposition...")
	fmt.Printf("   Task: %s\n\n", task)

	result, err := sv.Run(context.Background(), task)
	if err != nil {
		return fmt.Errorf("supervisor: %w", err)
	}

	fmt.Printf("\n🏁 Task Complete (%s)\n", result.Duration.Truncate(1e6))
	fmt.Printf("   Agents used: %d | Success: %d | Failed: %d\n",
		result.TaskCount, result.SuccessCount, result.FailureCount)
	fmt.Printf("\n📄 Final Output:\n%s\n", result.FinalOutput)

	return nil
}

func listAgents(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load(configDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	fmt.Println("Configured Agents:")
	fmt.Println(strings.Repeat("─", 60))
	for name, ac := range cfg.Agents {
		fmt.Printf("  %-12s  model=%-20s  provider=%-10s  tools=[%s]\n",
			name, ac.Model, ac.Provider, strings.Join(ac.Tools, ", "))
	}
	return nil
}

func listTools(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load(configDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	fmt.Println("Registered Tools:")
	fmt.Println(strings.Repeat("─", 60))
	for name, tc := range cfg.Tools {
		fmt.Printf("  %-15s  %s  (timeout: %s)\n", name, tc.Description, tc.Timeout)
	}
	return nil
}

func validateConfig(_ *cobra.Command, args []string) error {
	if err := config.Validate(args[0]); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	fmt.Println("✅ Configuration is valid")
	return nil
}

func createProvider(name string, pc config.ProviderConfig) provider.Provider {
	apiKey := os.Getenv(pc.APIKey)
	switch name {
	case "openai":
		if apiKey == "" {
			return nil
		}
		return provider.NewOpenAI(apiKey, pc.BaseURL)
	case "anthropic":
		if apiKey == "" {
			return nil
		}
		return provider.NewAnthropic(apiKey, pc.BaseURL)
	case "ollama":
		return provider.NewOllama(pc.BaseURL)
	default:
		return nil
	}
}

func resolveProvider(reg *provider.Registry, cfg *config.Config) provider.Provider {
	if providerFlag != "" {
		if p, ok := reg.Get(providerFlag); ok {
			return p
		}
	}
	for _, name := range []string{"openai", "anthropic", "ollama"} {
		if p, ok := reg.Get(name); ok {
			return p
		}
	}
	return nil
}

func resolveModel(cfg *config.Config) string {
	if modelFlag != "" {
		return modelFlag
	}
	for _, pc := range cfg.Providers {
		if pc.Model != "" {
			return pc.Model
		}
	}
	return "gpt-4o"
}

func buildFactory(p provider.Provider, mem *memory.Store, logger *slog.Logger) agent.AgentFactory {
	return func(agentType string) (*agent.Agent, error) {
		cfg := agent.Config{
			Name:          agentType,
			Model:         resolveModel(nil),
			MaxIterations: 10,
			TokenBudget:   16000,
		}
		switch agentType {
		case "researcher":
			return agents.NewResearcher(cfg, p, mem, logger), nil
		case "coder":
			return agents.NewCoder(cfg, p, mem, outputDir, logger), nil
		case "reviewer":
			return agents.NewReviewer(cfg, p, mem, logger), nil
		case "writer":
			return agents.NewWriter(cfg, p, mem, logger), nil
		default:
			reg := tools.NewRegistry()
			return agent.New(cfg, p, reg, logger), nil
		}
	}
}

func createDemoProvider(task string) *provider.MockProvider {
	return provider.NewMockProvider(
		&provider.Response{
			Content: fmt.Sprintf(`{
				"tasks": [
					{"id": "research", "description": "Research best practices for: %s", "agent": "researcher", "depends_on": []},
					{"id": "implement", "description": "Implement solution for: %s", "agent": "coder", "depends_on": ["research"]},
					{"id": "review", "description": "Review the implementation", "agent": "reviewer", "depends_on": ["implement"]},
					{"id": "document", "description": "Write documentation", "agent": "writer", "depends_on": ["review"]}
				]
			}`, task, task),
			Usage: provider.Usage{TotalTokens: 50},
		},
		&provider.Response{Content: "Research findings compiled.", Usage: provider.Usage{TotalTokens: 30}},
		&provider.Response{Content: "Implementation complete.", Usage: provider.Usage{TotalTokens: 30}},
		&provider.Response{Content: "Code review: approved.", Usage: provider.Usage{TotalTokens: 20}},
		&provider.Response{Content: "Documentation written.", Usage: provider.Usage{TotalTokens: 20}},
		&provider.Response{
			Content: fmt.Sprintf("Task completed successfully. All agents collaborated to deliver: %s", task),
			Usage:   provider.Usage{TotalTokens: 40},
		},
	)
}
