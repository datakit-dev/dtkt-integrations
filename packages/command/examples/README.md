# Examples

These examples depend on a running instance of the Command integration package.

## Connections

Command connections for use in examples.

### Command

Connect to the Command integration.

#### Local command execution (relative to the running integration)

```shell
dtkt create connection command -f examples/configs/local.json --intgr command
```

<img alt="connect-local demo with VHS" src="./connections/connect-local/vhs.gif" width="800" />

#### OR - Remote command execution (SSH)

```shell
dtkt create connection command -f examples/configs/remote.json --intgr command
```

<img alt="connect-remote demo with VHS" src="./connections/connect-remote/vhs.gif" width="800" />

## Services

### CommandService

#### ExecuteCommand

Execute a single command.

```shell
dtkt call ExecuteCommand \
  --conn command \
  -f examples/command-service/execute-command/input.json
```

<img alt="execute-command demo with VHS" src="./command-service/execute-command/vhs.gif" width="800" />

#### ExecuteStreamedCommand

Execute a single command with streaming input and output.

```shell
dtkt call ExecuteStreamedCommand \
  --conn command \
  -f examples/command-service/execute-streamed-command/inputs.jsonl
```

<img alt="execute-streamed-command demo with VHS" src="./command-service/execute-streamed-command/vhs.gif" width="800" />

#### ExecuteCommands

Execute a stream of commands.

```shell
dtkt call ExecuteCommands \
  --conn command \
  -f examples/command-service/execute-commands/inputs.jsonl
```

<img alt="execute-commands demo with VHS" src="./command-service/execute-commands/vhs.gif" width="800" />

#### ExecuteBatchCommands

Execute multiple commands as a batch.

```shell
dtkt call ExecuteBatchCommands \
  --conn command \
  -f examples/command-service/execute-batch-commands/input.json
```

<img alt="execute-batch-commands demo with VHS" src="./command-service/execute-batch-commands/vhs.gif" width="800" />

#### ExecuteShellCommand

Execute a single shell command.

```shell
dtkt call ExecuteShellCommand \
  --conn command \
  -f examples/command-service/execute-shell-command/input.json
```

<img alt="execute-shell-command demo with VHS" src="./command-service/execute-shell-command/vhs.gif" width="800" />

#### ExecuteStreamedShellCommand

Execute a single shell command with streaming input and output.

```shell
dtkt call ExecuteStreamedShellCommand \
  --conn command \
  -f examples/command-service/execute-streamed-shell-command/input.json
```

<img alt="execute-streamed-shell-command demo with VHS" src="./command-service/execute-streamed-shell-command/vhs.gif" width="800" />

#### ExecuteShellCommands

Execute a stream of shell commands.

```shell
dtkt call ExecuteShellCommands \
  --conn command \
  -f examples/command-service/execute-shell-commands/inputs.jsonl
```

<img alt="execute-shell-commands demo with VHS" src="./command-service/execute-shell-commands/vhs.gif" width="800" />

#### ExecuteBatchShellCommands

Execute multiple shell commands as a batch.

```shell
dtkt call ExecuteBatchShellCommands \
  --conn command \
  -f examples/command-service/execute-batch-shell-commands/input.json
```

<img alt="execute-batch-shell-commands demo with VHS" src="./command-service/execute-batch-shell-commands/vhs.gif" width="800" />

#### TerminalSession

Start a terminal session and stream raw input/output.

```shell
dtkt call TerminalSession \
  --conn command \
  -f examples/command-service/terminal-session/inputs.jsonl
```

<img alt="terminal-session demo with VHS" src="./command-service/terminal-session/vhs.gif" width="800" />

## Flows

### Coming soon!
