<div align="center">
  <img height="300" src="https://github.com/kokaq/.github/blob/main/kokaq-repl.png?raw=true" alt="cute quokka as kokaq logo"/>
</div>

`kokaq-repl` is an interactive REPL (Read-Eval-Print Loop) for exploring and debugging [kokaq](https://github.com/yourorg/kokaq) message queues via gRPC. `kokaq-repl` provides a developer-friendly CLI for inspecting namespaces, managing queues, and working with messages — including DLQ, peek-lock, and consumer group support.

[![Go Reference](https://pkg.go.dev/badge/github.com/kokaq/repl.svg)](https://pkg.go.dev/github.com/kokaq/repl)
[![Tests](https://github.com/kokaq/repl/actions/workflows/go.yml/badge.svg)](https://github.com/kokaq/repl/actions/workflows/go.yml)

## 📦 Installation

```bash
go install github.com/kokaq/repl@latest
```

## 🧪 Usage
```bash
kokaq-repl --server localhost:9000 --namespace default
```
Inside the shell:
```bash
kokaq> list namespaces
kokaq> create namespace dev
kokaq> create queue orders --visibility 30s
kokaq> enqueue orders "hello world"
kokaq> dequeue orders
kokaq> ack 8423bcd0-221b-49d0
kokaq> dlq reprocess orders
```

## ⚙️ CLI Options

```bash
--server         Address of the gRPC server (default: localhost:9000)
--namespace      Default namespace to use
--tls            Enable TLS
--token          Bearer auth token
--log-level      Log verbosity: debug | info | warn | error
```

## 🧠 Examples

```bash
# Create a queue and send messages
kokaq> create queue logs
kokaq> enqueue logs "log entry 1"
kokaq> enqueue logs "log entry 2"

# Receive and ack messages
kokaq> dequeue logs
kokaq> ack <message-id>

# Explore DLQ
kokaq> dlq list logs
kokaq> dlq reprocess logs --limit 10
```

## 🧱 Contributing

Contributions welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for code style and testing requirements.

## 📜 License

[MIT](./LICENSE) — open-source and production-ready.
