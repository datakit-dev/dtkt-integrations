# Examples

These examples depend on a running instance of the Fivetran integration package.

## Connections

Fivetran connections for use in examples.

### Fivetran

Create the fivetran connection.

```shell
dtkt conn create \
  --intgr fivetran \
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
### Echo Action

Run the echo action.

```shell
dtkt call ExecuteAction -d '{
  "name": "actions/echo",
  "input": {
    "message": "Hello, World!"
  }
}'
```

<img alt="echo-action demo with VHS" src="./echo-action/vhs.gif" width="800" />

## Flows
### Greet Flow

Run the greet flow.

```shell
cat ./greet-flow/inputs.jsonl | dtkt flow run ./greet-flow/flow.dtkt.yaml
```

<img alt="greet-flow demo with VHS" src="./greet-flow/vhs.gif" width="800" />

### Echo Flow

Run the echo flow.

```shell
cat ./echo-flow/inputs.jsonl | dtkt flow run ./echo-flow/flow.dtkt.yaml
```

<img alt="echo-flow demo with VHS" src="./echo-flow/vhs.gif" width="800" />
