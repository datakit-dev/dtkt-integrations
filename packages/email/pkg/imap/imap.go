package imap

import (
	"fmt"
	"log/slog"
	"mime"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/charset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"
	eventv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/event/v1beta1"

	emailintgr "github.com/datakit-dev/dtkt-integrations/email/pkg/proto/dtkt/emailintgr/v1beta1"
)

type ImapService struct {
	config        *emailintgr.Config_ImapConfig
	clientOptions *imapclient.Options

	events *v1beta1.EventRegistry
}

func NewImapService(config *emailintgr.Config_ImapConfig, events *v1beta1.EventRegistry) (*ImapService, error) {
	options := &imapclient.Options{
		WordDecoder: &mime.WordDecoder{CharsetReader: charset.Reader},
	}

	return &ImapService{
		config:        config,
		clientOptions: options,
		events:        events,
	}, nil
}

func (s *ImapService) CheckConfig() error {
	client, err := imapclient.DialTLS(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port), s.clientOptions)

	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to connect to IMAP server: %v", err))
	}

	if err := client.Login(s.config.Username, s.config.Password); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to authenticate to IMAP server: %v", err))
	}

	err = client.Close()
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to close IMAP connection: %v", err))
	}

	return nil
}

func parseImapAddress(addr imap.Address) *emailv1beta1.EmailAddress {
	return &emailv1beta1.EmailAddress{
		Name:    addr.Name,
		Address: addr.Addr(),
	}
}

func parseImapAddressList(addrList []imap.Address) []*emailv1beta1.EmailAddress {
	emailAddrs := make([]*emailv1beta1.EmailAddress, len(addrList))
	for i, addr := range addrList {
		emailAddrs[i] = parseImapAddress(addr)
	}
	return emailAddrs
}

// Note (jordan): While updating the event/event source API used in this package
// I noticed this method isn't called?
func (s *ImapService) GetEvents(stream grpc.ServerStreamingServer[eventv1beta1.StreamPullEventsResponse]) error {
	ctx := stream.Context()

	c, err := imapclient.DialTLS(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port), s.clientOptions)

	if err != nil {
		return fmt.Errorf("failed to connect to IMAP server: %v", err)
	}

	defer func() {
		if err := c.Close(); err != nil {
			log.Error(ctx, "failed to close IMAP connection", log.Err(err))
		}
	}()

	if err := c.Login(s.config.Username, s.config.Password); err != nil {
		return fmt.Errorf("failed to authenticate to IMAP server: %v", err)
	}

	// List mailboxes
	mailboxes, err := c.List("", "%", nil).Collect()
	if err != nil {
		return fmt.Errorf("failed to list mailboxes: %v", err)
	}

	log.Debug(ctx, "mailboxes", slog.Any("mailboxes", mailboxes))

	selectedMbox, err := c.Select("INBOX", nil).Wait()
	if err != nil {
		return fmt.Errorf("failed to select INBOX: %v", err)
	}
	log.Debug(ctx, "selected mailbox", slog.Any("mailbox", selectedMbox))

	var messages []*imapclient.FetchMessageBuffer
	if selectedMbox.NumMessages > 0 {
		seqSet := imap.SeqSetNum(1)
		fetchOptions := &imap.FetchOptions{Envelope: true}
		messages, err = c.Fetch(seqSet, fetchOptions).Collect()
		if err != nil {
			return fmt.Errorf("failed to fetch first message in INBOX: %v", err)
		}
		log.Debug(ctx, "subject of first message in INBOX", slog.Any("subject", messages[0].Envelope.Subject))
	}

	for _, msg := range messages {
		event, err := s.events.Find("EmailReceived")
		if err != nil {
			return fmt.Errorf("failed to find registered event %s: %w", "EmailReceived", err)
		}

		var textBody []byte
		var htmlBody []byte

		for _, body := range msg.BodySection {
			log.Debug(ctx, "body", slog.Any("bodyHeader", body.Section.HeaderFields), slog.Any("body", string(body.Bytes)), slog.Any("bodySpecifier", body.Section.Specifier))
		}

		dateTime := msg.Envelope.Date

		from := parseImapAddressList(msg.Envelope.From)[0]
		returnPath := parseImapAddressList(msg.Envelope.Sender)[0]
		// replyTo := parseImapAddressList(msg.Envelope.ReplyTo)
		to := parseImapAddressList(msg.Envelope.To)
		cc := parseImapAddressList(msg.Envelope.Cc)
		bcc := parseImapAddressList(msg.Envelope.Bcc)

		payload, err := anypb.New(&emailv1beta1.ReceivedEmail{
			Id: msg.Envelope.MessageID,
			At: dateTime.Format(time.RFC3339),
			Email: &emailv1beta1.Email{
				From:       from,
				ReturnPath: returnPath,
				To:         to,
				Cc:         cc,
				Bcc:        bcc,
				Subject:    msg.Envelope.Subject,
				TextBody:   string(textBody),
				HtmlBody:   string(htmlBody),
			},
		})
		if err != nil {
			log.Error(ctx, "failed to set payload", log.Err(err))
		}

		err = stream.Send(&eventv1beta1.StreamPullEventsResponse{
			// EventSource: ,
			Event:   event.Proto().GetName(),
			Payload: payload,
		})
		if err != nil {
			log.Error(ctx, "failed to send event", log.Err(err))
		}
	}

	if err := c.Logout().Wait(); err != nil {
		return fmt.Errorf("failed to logout from IMAP server: %v", err)
	}

	return nil
}
