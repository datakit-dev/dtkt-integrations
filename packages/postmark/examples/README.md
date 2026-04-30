# Examples

These examples depend on a running instance of the Postmark integration package.
Unless you are running a Postmark sandbox server, these emails will get sent to real email addresses.

The `to` user in these examples is set to `shadi@withdatakit.com` our CTO.
He'll be happy to receive your emails, but please don't spam him! 😉

## Connections

Postmark connections for use in examples.

### Postmark

Connect to the Postmark integration.

```shell
ACCOUNT_API_KEY=replace SERVER_API_KEY=me envsubst '$ACCOUNT_API_KEY,$SERVER_API_KEY' < examples/configs/postmark.envsubst.json > examples/configs/postmark.json
dtkt create connection postmark -f examples/configs/postmark.json --intgr postmark
```

<img alt="connect-postmark demo with VHS" src="./connections/connect-postmark/vhs.gif" width="800" />

## Services

### EmailService

#### SendEmail

Send a single email.

```shell
dtkt call SendEmail \
  --conn postmark \
  -f examples/email-service/send-email/input.json
```

<img alt="send-email demo with VHS" src="./email-service/send-email/vhs.gif" width="800" />

#### SendEmails

Send a stream of emails.

```shell
dtkt call SendEmails \
  --conn postmark \
  -f examples/email-service/send-emails/inputs.jsonl
```

<img alt="send-emails demo with VHS" src="./email-service/send-emails/vhs.gif" width="800" />

#### SendBatchEmail

Send multiple emails as a batch.

```shell
dtkt call SendBatchEmail \
  --conn postmark \
  -f examples/email-service/send-batch-email/input.json
```

<img alt="send-batch-email demo with VHS" src="./email-service/send-batch-email/vhs.gif" width="800" />

#### SendEmailWithTemplate

Send a single email using a template.

```shell
dtkt call SendEmailWithTemplate \
  --conn postmark \
  -f examples/email-service/send-email-with-template/input.json
```

<img alt="send-email-with-template demo with VHS" src="./email-service/send-email-with-template/vhs.gif" width="800" />

#### SendEmailsWithTemplate

Send a stream of emails using a template.

```shell
dtkt call SendEmailsWithTemplate \
  --conn postmark \
  -f examples/email-service/send-emails-with-template/inputs.jsonl
```

<img alt="send-emails-with-template demo with VHS" src="./email-service/send-emails-with-template/vhs.gif" width="800" />

#### SendBatchEmailWithTemplate

Send multiple emails using a template as a batch.

```shell
dtkt call SendBatchEmailWithTemplate \
  --conn postmark \
  -f examples/email-service/send-batch-email-with-template/input.json
```

<img alt="send-batch-email-with-template demo with VHS" src="./email-service/send-batch-email-with-template/vhs.gif" width="800" />

#### ListEmailTemplates

List all email templates.

```shell
dtkt call ListEmailTemplates \
  --conn postmark \
  -f examples/email-service/list-email-templates/input.json
```

<img alt="list-email-templates demo with VHS" src="./email-service/list-email-templates/vhs.gif" width="800" />

#### GetEmailTemplate

Get an email template by its ID (this is mapped to the `Alias` field in Postmark).

```shell
dtkt call GetEmailTemplate \
  --conn postmark \
  -f examples/email-service/get-email-template/input.json
```

<img alt="get-email-template demo with VHS" src="./email-service/get-email-template/vhs.gif" width="800" />

#### CreateEmailTemplate

Create an email template.

```shell
dtkt call CreateEmailTemplate \
  --conn postmark \
  -f examples/email-service/create-email-template/input.json
```

<img alt="create-email-template demo with VHS" src="./email-service/create-email-template/vhs.gif" width="800" />

#### UpdateEmailTemplate

Update an email template.

```shell
dtkt call UpdateEmailTemplate \
  --conn postmark \
  -f examples/email-service/update-email-template/input.json
```

<img alt="update-email-template demo with VHS" src="./email-service/update-email-template/vhs.gif" width="800" />

#### DeleteEmailTemplate

Delete an email template.

```shell
dtkt call DeleteEmailTemplate \
  --conn postmark \
  -f examples/email-service/delete-email-template/input.json
```

<img alt="delete-email-template demo with VHS" src="./email-service/delete-email-template/vhs.gif" width="800" />

### ActionService

#### ListActions

List all available actions.

```shell
dtkt call ListActions \
  --conn postmark \
  -f examples/action-service/list-actions/input.json
```

<img alt="list-actions demo with VHS" src="./action-service/list-actions/vhs.gif" width="800" />

#### GetAction

Get details of a specific action.

```shell
dtkt call GetAction \
  --conn postmark \
  -f examples/action-service/get-action/input.json
```

<img alt="get-action demo with VHS" src="./action-service/get-action/vhs.gif" width="800" />

#### ExecuteAction

##### Example

Execute an example action.

```shell
dtkt call ExecuteAction \
  --conn postmark \
  -f examples/action-service/execute-action/example/input.json
```

<img alt="execute-action-example demo with VHS" src="./action-service/execute-action/example/vhs.gif" width="800" />
