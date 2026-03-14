# Postmark Integration

The `Postmark` integration implements:

- `ActionService`
- `BaseService`
- `EmailService`

Send individual, batched, streamed, or templated emails through Postmark's API.

<img alt="send-email demo with VHS" src="./examples/email-service/send-email/vhs.gif" width="800" />

## Getting Started

Download or build the [DataKit CLI](https://withdatakit.com/docs/cli/dtkt/overview) and ensure it is accessible to your `$PATH`.

```shell
dtkt version
```

### Start the Package

You can start the integration in one of two ways:

#### Option 1: Develop from Source

Clone down this repository and the [DataKit SDK](https://github.com/datakit-dev/dtkt-sdk) side by side.

From the root of the integration package, run:

```shell
dtkt intgr dev
```

This launches the integration in a development loop that automatically rebuilds and restarts on detected changes.

<img alt="dev demo with VHS" src="./examples/dev/vhs.gif" width="800" />

#### Option 2: Run a prebuilt binary in Docker

Ensure the Docker daemon is running and run the following command:

```shell
dtkt intgr start postmark --runtime docker
```

This launches the integration in a Docker container.

### Ensure your integration is registered

In another terminal, run:

```shell
dtkt intgr get postmark
```

### Configure the integration

There are sample configurations in [`examples/configs`](./examples/configs).

Create the config file with your Postmark API tokens:

```shell
envsubst '$ACCOUNT_API_KEY,$SERVER_API_KEY' < examples/configs/postmark.envsubst.json > examples/configs/postmark.json
```

If you'd like to create your own, refer to the Config struct in [pkg/instance.go](./pkg/instance.go) or simply create one interactively when creating a connection.

## Examples

Refer to the [examples](./examples/README.md) for more detailed examples of how to use the `Postmark` integration package.

## Limitations

The only template language supported in this integration is Postmark's Mustachio template syntax. Refer to the [Postmark documentation](https://postmarkapp.com/support/article/1077-template-syntax) for more information on the syntax.

## Related Packages

- [Email](../email): Send individual, batched, or streamed emails to any compatible SMTP server and subscribe to email events through POP3 and IMAP.
- [Mailpit](../mailpit): A simple and easy-to-use email testing tool that allows developers to send, receive, and view emails in a development environment.
