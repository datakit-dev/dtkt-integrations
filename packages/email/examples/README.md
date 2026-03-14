# Examples

These examples depend on a running instance of the Email and [Mailpit](../../mailpit/README.md) integration packages.

## Connections

Email connections for use in examples.

### Mailpit

Connect to the Email integration using Mailpit.

```shell
dtkt connect create mailpit -f examples/configs/mailpit.json --intgr email
```

<img alt="connect-mailpit demo with VHS" src="./connections/connect-mailpit/vhs.gif" width="800" />

## Services

### EmailService

#### SendEmail

Send a single email.

```shell
dtkt call SendEmail \
  --conn mailpit \
  -f examples/email-service/send-email/input.json
```

<img alt="send-email demo with VHS" src="./email-service/send-email/vhs.gif" width="800" />

#### SendEmails

Send a stream of emails.

```shell
dtkt call SendEmails \
  --conn mailpit \
  -f examples/email-service/send-emails/inputs.jsonl
```

<img alt="send-emails demo with VHS" src="./email-service/send-emails/vhs.gif" width="800" />

#### SendBatchEmail

Send multiple emails as a batch.

```shell
dtkt call SendBatchEmail \
  --conn mailpit \
  -f examples/email-service/send-batch-email/input.json
```

<img alt="send-batch-email demo with VHS" src="./email-service/send-batch-email/vhs.gif" width="800" />

#### SendEmailWithTemplate

Send a single email using a template.

```shell
dtkt call SendEmailWithTemplate \
  --conn mailpit \
  -f examples/email-service/send-email-with-template/input.json
```

<img alt="send-email-with-template demo with VHS" src="./email-service/send-email-with-template/vhs.gif" width="800" />

#### SendEmailsWithTemplate

Send a stream of emails using a template.

```shell
dtkt call SendEmailsWithTemplate \
  --conn mailpit \
  -f examples/email-service/send-emails-with-template/inputs.jsonl
```

<img alt="send-emails-with-template demo with VHS" src="./email-service/send-emails-with-template/vhs.gif" width="800" />

#### SendBatchEmailWithTemplate

Send multiple emails using a template as a batch.

```shell
dtkt call SendBatchEmailWithTemplate \
  --conn mailpit \
  -f examples/email-service/send-batch-email-with-template/input.json
```

<img alt="send-batch-email-with-template demo with VHS" src="./email-service/send-batch-email-with-template/vhs.gif" width="800" />

#### ListEmailTemplates

List all email templates.

```shell
dtkt call ListEmailTemplates \
  --conn mailpit \
  -f examples/email-service/list-email-templates/input.json
```

<img alt="list-email-templates demo with VHS" src="./email-service/list-email-templates/vhs.gif" width="800" />

#### GetEmailTemplate

Get an email template.

```shell
dtkt call GetEmailTemplate \
  --conn mailpit \
  -f examples/email-service/get-email-template/input.json
```

<img alt="get-email-template demo with VHS" src="./email-service/get-email-template/vhs.gif" width="800" />

#### CreateEmailTemplate

Create an email template.

```shell
dtkt call CreateEmailTemplate \
  --conn mailpit \
  -f examples/email-service/create-email-template/input.json
```

<img alt="create-email-template demo with VHS" src="./email-service/create-email-template/vhs.gif" width="800" />

#### UpdateEmailTemplate

Update an email template.

```shell
dtkt call UpdateEmailTemplate \
  --conn mailpit \
  -f examples/email-service/update-email-template/input.json
```

<img alt="update-email-template demo with VHS" src="./email-service/update-email-template/vhs.gif" width="800" />

#### DeleteEmailTemplate

Delete an email template.

```shell
dtkt call DeleteEmailTemplate \
  --conn mailpit \
  -f examples/email-service/delete-email-template/input.json
```

<img alt="delete-email-template demo with VHS" src="./email-service/delete-email-template/vhs.gif" width="800" />

## Flows

### Summarize Emails

This example flow subscribes to an inbox and summarizes the emails received.

```shell
dtkt flow run ./summarize-emails/flow.dtkt.yaml
```

<img alt="summarize-emails demo with VHS" src="./summarize-emails/vhs.gif" width="800" />

### Auto Reply

This example flow automatically writes a reply to emails received in an inbox.
It will prompt the user to confirm the reply before sending it.

```shell
dtkt flow run ./auto-reply/flow.dtkt.yaml
```

<img alt="auto-reply demo with VHS" src="./auto-reply/vhs.gif" width="800" />
