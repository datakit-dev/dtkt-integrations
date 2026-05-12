# Examples

These examples depend on a running instance of the Chrome integration package.

## Connections

Chrome connections for use in examples.

### Chrome

Create the chrome connection.

```shell
dtkt conn create \
  --intgr chrome \
  --config-path examples/config.json
```

<img alt="connection demo with VHS" src="./connection/vhs.gif" width="800" />

## Services
### GreetService.Greet

Run the Greet RPC.

```shell
dtkt call Greet -d '{"name": "World"}'
```

<img alt="greet-rpc demo with VHS" src="./greet-rpc/vhs.gif" width="800" />

## Actions
No example actions.

## Flows
### Greet Flow

Run the greet flow.

```shell
cat ./greet-flow/inputs.jsonl | dtkt flow run ./greet-flow/flow.dtkt.yaml
```

<img alt="greet-flow demo with VHS" src="./greet-flow/vhs.gif" width="800" />
