package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	prompt "github.com/c-bata/go-prompt"
	"github.com/kokaq/client/sdk"
	"github.com/kokaq/protocol/proto"
)

type ReplClient struct {
	address          string
	currentNamespace string
	currentQueue     string
	ctx              context.Context
	client           *sdk.KokaqClient
}

func NewReplClient(address string) *ReplClient {

	var ctx context.Context = context.Background()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nExiting Kokaq REPL...")
		os.Exit(0)
	}()
	return &ReplClient{
		address: address,
		ctx:     ctx,
	}
}

func (repl *ReplClient) promptPrefix() string {
	if repl.currentNamespace != "" && repl.currentQueue != "" {
		return fmt.Sprintf("kokaq [ns: %s] [q: %s] > ", repl.currentNamespace, repl.currentQueue)
	} else if repl.currentNamespace != "" && repl.currentQueue == "" {
		return fmt.Sprintf("kokaq [ns: %s] > ", repl.currentNamespace)
	} else {
		return "kokaq > "
	}

}

func (repl *ReplClient) getClient() (*sdk.KokaqClient, error) {
	var opts = &sdk.KokaqClientOptions{
		TLSEnabled:  false,
		DialTimeout: 5 * time.Second,
	}
	return sdk.NewKokaqClient(repl.address, opts)

}

func (repl *ReplClient) Start() {
	fmt.Println(`
██╗  ██╗ ██████╗ ██╗  ██╗ █████╗ ███████╗
██║ ██╔╝██╔═══██╗██║ ██╔╝██╔══██╗██╔══██║
█████╔╝ ██║   ██║█████╔╝ ███████║██║  ██║
██╔═██╗ ██║   ██║██╔═██╗ ██╔══██║██║  ██║
██║  ██╗╚██████╔╝██║  ██╗██║  ██║██████╔╝
╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚════██╗
                                      ╚═╝
Welcome to Kokaq REPL — Type 'help' for commands. Type 'exit' to quit.
`)
	var err error
	repl.client, err = repl.getClient()
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		os.Exit(1)
	}
	defer repl.client.Close()

	p := prompt.New(
		repl.executor,
		repl.completer,
		prompt.OptionPrefix(repl.promptPrefix()),
		prompt.OptionTitle("kokaq-repl"),
	)
	p.Run()
}

func (repl *ReplClient) completer(d prompt.Document) []prompt.Suggest {
	// Simplified example. Add more as needed
	return prompt.FilterHasPrefix([]prompt.Suggest{
		{Text: "namespace create", Description: "Create a namespace"},
		{Text: "namespace delete", Description: "Delete an existing namespace"},
		{Text: "namespace list", Description: "List all namespaces"},
		{Text: "namespace use", Description: "Set current namespace"},
		{Text: "queue create", Description: "Create a new queue in current namespace"},
		{Text: "queue delete", Description: "Delete a queue from current namespace"},
		{Text: "queue list", Description: "List queues in selected namespace"},
		{Text: "queue use", Description: "Set current queue"},
		{Text: "enqueue", Description: "Enqueue message with 64-bit priority"},
		{Text: "dequeue", Description: "Dequeue highest-priority message"},
		{Text: "ack", Description: "Acknowledge successful message"},
		{Text: "nack", Description: "Mark message as failed"},
		{Text: "peek", Description: "Peek highest-priority message"},
		{Text: "help", Description: "Show help guide"},
	}, d.GetWordBeforeCursor(), true)
}

func (repl *ReplClient) executor(in string) {

	args := strings.Fields(in)
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "help":
		fmt.Println(`kokaq repl help guide

This interactive shell lets you manage namespaces and queues, 
and perform message operations on a distributed priority queue system.

Namespace Commands
  namespace create <namespace>   Create a new namespace
  namespace delete <namespace>   Delete an existing namespace
  namespace list                 List all namespaces
  namespace use <namespace>      Set current namespace

Queue Commands
  queue create <queue>           Create a new queue in current namespace
  queue delete <queue>           Delete a queue from current namespace
  queue list                     List queues in selected namespace
  queue use <queue>              Set current queue

Message Commands
  enqueue <message> <priority>   Enqueue message with 64-bit priority
  dequeue                        Dequeue highest-priority message
  ack <messageID> <lockID>       Acknowledge successful message
  nack <messageID> <lockID>      Mark message as failed
  peek                           Peek highest-priority message

Utility
  help                           Show help menu

Notes
- You must first 'namespace use' and then 'queue use' before message operations.
- Press Ctrl+C to cancel ongoing operations.`)
		os.Exit(0)
	case "exit":
		fmt.Println("Bye!")
		os.Exit(0)
	case "namespace":
		repl.handleNamespace(args[1:])
	case "queue":
		repl.handleQueue(args[1:])
	case "enqueue":
		repl.handleEnqueue(args[1:])
	case "dequeue":
		repl.handleDequeue()
	case "ack":
		repl.handleAck(args[1:])
	case "nack":
		repl.handleNack(args[1:])
	case "peek":
		repl.handlePeek()
	default:
		fmt.Println("Unknown command:", in)
	}
}

func (repl *ReplClient) handleNamespace(args []string) {
	if len(args) == 0 {
		fmt.Println("namespace: missing subcommand")
		return
	}
	var currentNamespace string
	switch args[0] {
	case "create":
		if len(args) < 2 {
			fmt.Println("usage: namespace use <name>")
			return
		}
		currentNamespace = args[1]
		resp, err := repl.client.CreateNamespace(repl.ctx, currentNamespace)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Created namespace:", resp.Namespace)
		repl.currentNamespace = currentNamespace
		fmt.Println("Selected namespace:", args[1])
	case "delete":
		if len(args) < 2 {
			fmt.Println("usage: namespace use <name>")
			return
		}
		currentNamespace = args[1]
		nc, err := repl.client.GetNamespaceClient(repl.ctx, currentNamespace)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if err := nc.Delete(repl.ctx); err != nil {
			fmt.Println("Error: cannot delete namespace:", err)
			return
		}
		repl.currentNamespace = ""
		fmt.Println("Deleted namespace:", args[1])
	case "use":
		if len(args) < 2 {
			fmt.Println("usage: namespace use <name>")
			return
		}
		currentNamespace = args[1]
		_, err := repl.client.GetNamespaceClient(repl.ctx, currentNamespace)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		repl.currentNamespace = currentNamespace
		fmt.Println("Selected namespace:", args[1])
	case "list":
		fmt.Println("Namespaces: ns1, ns2")
	default:
		fmt.Println("namespace: unknown subcommand")
	}
}

func (repl *ReplClient) handleQueue(args []string) {
	if len(args) == 0 {
		fmt.Println("queue: missing subcommand")
		return
	}
	if repl.currentNamespace == "" {
		fmt.Println("Set namespace first using 'namespace use'")
		return
	}
	var currentQueue string
	switch args[0] {
	case "create":
		if len(args) < 2 {
			fmt.Println("usage: queue create <name>")
			return
		}
		currentQueue = args[1]
		_, err := repl.client.CreateQueue(repl.ctx, repl.currentNamespace, currentQueue)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		repl.currentQueue = currentQueue
		fmt.Println("Created queue:", repl.currentQueue)
		fmt.Println("Selected queue:", repl.currentQueue)
	case "delete":
		if len(args) < 2 {
			fmt.Println("usage: namespace use <name>")
			return
		}
		currentQueue = args[1]
		nc, err := repl.client.GetQueueClient(repl.ctx, repl.currentNamespace, currentQueue)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if err := nc.Delete(repl.ctx); err != nil {
			fmt.Println("Error: cannot delete namespace:", err)
			return
		}
		repl.currentQueue = ""
		fmt.Println("Deleted namespace:", args[1])
	case "use":
		if len(args) < 2 {
			fmt.Println("usage: queue use <name>")
			return
		}
		repl.currentQueue = args[1]

		currentQueue = args[1]
		_, err := repl.client.GetQueueClient(repl.ctx, repl.currentNamespace, currentQueue)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		repl.currentQueue = currentQueue
		fmt.Println("Selected queue:", repl.currentQueue)
	case "list":
		fmt.Println("Queues: q1, q2")
	default:
		fmt.Println("queue: unknown subcommand")
	}
}

func (repl *ReplClient) handleEnqueue(args []string) {
	if repl.currentNamespace == "" || repl.currentQueue == "" {
		fmt.Println("Set namespace and queue first using 'namespace use' and 'queue use'")
		return
	}
	if len(args) < 3 || args[len(args)-2] != "priority" {
		fmt.Println("usage: enqueue <message> priority <uint64>")
		return
	}
	msg := strings.Join(args[:len(args)-2], " ")
	var prio uint64
	_, err := fmt.Sscanf(args[len(args)-1], "%d", &prio)
	if err != nil {
		fmt.Println("Invalid priority value")
		return
	}
	qc, err := repl.client.GetQueueClient(repl.ctx, repl.currentNamespace, repl.currentQueue)
	if err != nil {
		fmt.Println("Error: cannot enqueue message")
	}
	if msgId, _, err := qc.Sender().Send(repl.ctx, []byte(msg), prio); err != nil {
		fmt.Println("enqueue error:", err)
	} else {
		fmt.Printf("Eequeued message: %s, Priority: %d\n", msgId, prio)
	}
}

func (repl *ReplClient) handleDequeue() {
	if repl.currentNamespace == "" || repl.currentQueue == "" {
		fmt.Println("Set namespace and queue first using 'namespace use' and 'queue use'")
		return
	}
	qc, err := repl.client.GetQueueClient(repl.ctx, repl.currentNamespace, repl.currentQueue)
	if err != nil {
		fmt.Println("Error: cannot enqueue message")
	}
	if err := qc.Receiver().PollAndProcess(repl.ctx, 1, func(ctx context.Context, msgs []*proto.KokaqMessageResponse) error {
		for _, msg := range msgs {
			fmt.Printf("Dequeued message: %s, Priority: %d\n", msg.Message.MessageId, msg.Message.Priority)
		}
		return nil
	}); err != nil {
		fmt.Println("dequeue error:", err)
	}
}

func (repl *ReplClient) handleAck(args []string) {
	if len(args) < 2 {
		fmt.Println("usage: ack <message_id> <lock_id>")
		return
	}
	if repl.currentNamespace == "" || repl.currentQueue == "" {
		fmt.Println("Set namespace and queue first using 'namespace use' and 'queue use'")
		return
	}
	qc, err := repl.client.GetQueueClient(repl.ctx, repl.currentNamespace, repl.currentQueue)
	if err != nil {
		fmt.Println("Error: cannot ack message")
	}
	if err := qc.Receiver().Ack(repl.ctx, args[0], args[1]); err != nil {
		fmt.Println("ack error:", err)
	} else {
		fmt.Println("ack done")
	}
}

func (repl *ReplClient) handleNack(args []string) {
	if len(args) < 2 {
		fmt.Println("usage: nack <message_id> <lock_id>")
		return
	}
	if repl.currentNamespace == "" || repl.currentQueue == "" {
		fmt.Println("Set namespace and queue first using 'namespace use' and 'queue use'")
		return
	}
	qc, err := repl.client.GetQueueClient(repl.ctx, repl.currentNamespace, repl.currentQueue)
	if err != nil {
		fmt.Println("Error: cannot ack message")
	}
	if err := qc.Receiver().Nack(repl.ctx, args[0], args[1], proto.FailureReason_PROCESSING_ERROR, true); err != nil {
		fmt.Println("nack error:", err)
	} else {
		fmt.Println("nack done")
	}
}

func (repl *ReplClient) handlePeek() {
	if repl.currentNamespace == "" || repl.currentQueue == "" {
		fmt.Println("Set namespace and queue first using 'namespace use' and 'queue use'")
		return
	}
	qc, err := repl.client.GetQueueClient(repl.ctx, repl.currentNamespace, repl.currentQueue)
	if err != nil {
		fmt.Println("Error: cannot peek message")
	}
	if err := qc.Receiver().PeekAndProcess(repl.ctx, 1, 0, func(ctx context.Context, msg *proto.KokaqMessageResponse, lockID string) error {
		fmt.Printf("Peeked message: %s, Priority: %d\n", msg.Message.MessageId, msg.Message.Priority)
		return nil
	}); err != nil {
		fmt.Println("peek error:", err)
	}
}
