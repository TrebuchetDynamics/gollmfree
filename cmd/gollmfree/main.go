// Command gollmfree is a CLI for free anonymous LLM providers.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/gollmfree"
	"github.com/TrebuchetDynamics/gollmfree/providers"
)

const usageText = `gollmfree <command> [flags] [args]

Commands:
  chat    Send a message and print the reply
  list    Print registered providers and health summary
  models  Print model aliases and provider coverage

Run 'gollmfree <command> -help' for per-command flags.
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usageText)
		return 2
	}
	switch args[0] {
	case "chat":
		return cmdChat(args[1:])
	case "list":
		return cmdList(args[1:])
	case "models":
		return cmdModels(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "gollmfree: unknown command %q\n\n%s", args[0], usageText)
		return 2
	}
}

func cmdChat(args []string) int {
	fs := flag.NewFlagSet("chat", flag.ContinueOnError)
	model := fs.String("model", "auto", "model alias or provider name")
	timeout := fs.Duration("timeout", 60*time.Second, "per-attempt timeout")
	race := fs.Bool("race", false, "race top providers and use the fastest")
	_ = fs.Bool("stream", false, "stream output (emulated; output is still buffered)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "gollmfree chat: prompt argument required")
		fs.Usage()
		return 2
	}
	prompt := strings.Join(fs.Args(), " ")

	client := newDefaultClient(*timeout, *race)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout*3)
	defer cancel()
	resp, err := client.ChatCompletion(ctx, gollmfree.ChatRequest{
		Model:    *model,
		Messages: []gollmfree.Message{{Role: "user", Content: prompt}},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "gollmfree: %v\n", err)
		return 1
	}
	if len(resp.Choices) > 0 {
		text := resp.Choices[0].Message.Content
		fmt.Print(text)
		if !strings.HasSuffix(text, "\n") {
			fmt.Println()
		}
	}
	return 0
}

func cmdList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "output JSON")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	reg := defaultRegistry()
	if *asJSON {
		infos := reg.Providers()
		_ = json.NewEncoder(os.Stdout).Encode(infos)
		return 0
	}
	for _, p := range reg.Providers() {
		fmt.Printf("%-16s  priority=%-3d  models=%s\n",
			p.Name, p.DefaultPriority, strings.Join(p.SupportedModels, ","))
	}
	return 0
}

func cmdModels(args []string) int {
	fs := flag.NewFlagSet("models", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "output JSON")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	reg := defaultRegistry()
	if *asJSON {
		_ = json.NewEncoder(os.Stdout).Encode(reg.Models())
		return 0
	}
	for _, m := range reg.Models() {
		fmt.Printf("%-30s  providers=%s\n", m.Alias, strings.Join(m.Providers, ","))
	}
	return 0
}

func defaultRegistry() *gollmfree.Registry {
	pollinations := providers.NewPollinationsAI()
	chatai := providers.NewChatai()
	yqcloud := providers.NewYqcloud()
	wewordle := providers.NewWeWordle()
	reg, err := gollmfree.NewRegistry(
		gollmfree.ProviderInfo{Name: pollinations.Name(), Provider: pollinations, SupportedModels: pollinations.SupportedModels(), DefaultPriority: 1},
		gollmfree.ProviderInfo{Name: chatai.Name(), Provider: chatai, SupportedModels: chatai.SupportedModels(), DefaultPriority: 2},
		gollmfree.ProviderInfo{Name: yqcloud.Name(), Provider: yqcloud, SupportedModels: yqcloud.SupportedModels(), DefaultPriority: 3},
		gollmfree.ProviderInfo{Name: wewordle.Name(), Provider: wewordle, SupportedModels: wewordle.SupportedModels(), DefaultPriority: 4},
	)
	if err != nil {
		reg, _ = gollmfree.NewRegistry(
			gollmfree.ProviderInfo{Name: pollinations.Name(), Provider: pollinations, SupportedModels: pollinations.SupportedModels(), DefaultPriority: 1},
		)
	}
	return reg
}

func newDefaultClient(perAttemptTimeout time.Duration, race bool) *gollmfree.Client {
	opts := []gollmfree.Option{
		gollmfree.WithRegistry(defaultRegistry()),
		gollmfree.WithTimeout(perAttemptTimeout),
		gollmfree.WithMaxRetries(1),
	}
	if race {
		opts = append(opts, gollmfree.WithRaceMode(true))
	}
	return gollmfree.NewClient(opts...)
}
