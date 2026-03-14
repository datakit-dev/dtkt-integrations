# Examples

These examples depend on a running instance of the LocalBlob integration package.

## Connections

LocalBlob connections for use in examples.

### LocalBlob

Create the localblob connection.

```shell
dtkt conn create \
  --intgr localblob \
  --config-path examples/config.json
```

<img alt="example-connection demo with VHS" src="./example-connection/vhs.gif" width="800" />

## Actions

### Example Action

Run an example action.

```shell
dtkt action execute-action \
  --conn localblob \
  --name actions/example-echo \
  -f examples/example-echo-action/request.json
```

<img alt="example-echo-action demo with VHS" src="./example-echo-action/vhs.gif" width="800" />

## Flows

### Example Echo Flow

An example flow.

```shell
cat ./example-echo-flow/inputs.jsonl | jq -c | dtkt flow run ./example-echo-flow/flow.dtkt.yaml
```

<img alt="example-echo-flow demo with VHS" src="./example-echo-flow/vhs.gif" width="800" />
