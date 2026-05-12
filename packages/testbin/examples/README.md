# Examples

These examples depend on a running instance of the Testbin integration package.

## Connections

### Testbin

Connect to the Testbin integration. The connection's config carries a free-form `payload` that round-trips through `InspectService.GetConfig`.

```shell
dtkt create connection testbin -f examples/configs/testbin.json --intgr testbin
```

<img alt="connect demo with VHS" src="./connections/connect/vhs.gif" width="800" />

## Services

### EchoService

#### EchoUnary

Echo a payload back as a single response. Set `error` to fail the RPC instead.

```shell
dtkt call EchoUnary \
  --conn testbin \
  -f examples/echo-service/echo-unary/input.json
```

<img alt="echo-unary demo with VHS" src="./echo-service/echo-unary/vhs.gif" width="800" />

#### EchoServerStream

Emit `count` copies of the payload at `interval` spacing. Set `error_on` together with `error` to fail at a specific sequence index.

```shell
dtkt call EchoServerStream \
  --conn testbin \
  -f examples/echo-service/echo-server-stream/input.json
```

<img alt="echo-server-stream demo with VHS" src="./echo-service/echo-server-stream/vhs.gif" width="800" />

#### EchoClientStream

Stream payloads from the client; the server returns the total count and the last-seen payload.

```shell
dtkt call EchoClientStream \
  --conn testbin \
  -f examples/echo-service/echo-client-stream/inputs.jsonl
```

<img alt="echo-client-stream demo with VHS" src="./echo-service/echo-client-stream/vhs.gif" width="800" />

#### EchoBidiStream

For each incoming request, emit `count` copies of its payload. Mix in error-bearing requests to fail mid-stream.

```shell
dtkt call EchoBidiStream \
  --conn testbin \
  -f examples/echo-service/echo-bidi-stream/inputs.jsonl
```

<img alt="echo-bidi-stream demo with VHS" src="./echo-service/echo-bidi-stream/vhs.gif" width="800" />

### InspectService

#### Anything

Reflect back what the server saw on the incoming RPC: method, metadata, deadline, peer, and the request payload.

```shell
dtkt call Anything \
  --conn testbin \
  -f examples/inspect-service/anything/input.json
```

<img alt="anything demo with VHS" src="./inspect-service/anything/vhs.gif" width="800" />

#### GetConfig

Return the Config the integration instance was constructed with - the verbatim payload set on the connection.

```shell
dtkt call GetConfig \
  --conn testbin \
  -f examples/inspect-service/get-config/input.json
```

<img alt="get-config demo with VHS" src="./inspect-service/get-config/vhs.gif" width="800" />

## Flows

### Coming soon!
