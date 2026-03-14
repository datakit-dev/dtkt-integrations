package pop3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/knadh/go-pop3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"

	emailintgr "github.com/datakit-dev/dtkt-integrations/email/pkg/proto/dtkt/emailintgr/v1beta1"
)

type Pop3Service struct {
	config *emailintgr.Config_Pop3Config
	client *pop3.Client
}

func NewPop3Service(config *emailintgr.Config_Pop3Config) (*Pop3Service, error) {
	client := pop3.New(pop3.Opt{
		Host:       config.Host,
		Port:       int(config.Port),
		TLSEnabled: false,
	})

	return &Pop3Service{
		config: config,
		client: client,
	}, nil
}

func (s *Pop3Service) CheckConfig() error {
	client, err := s.client.NewConn()

	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to connect to POP3 server: %v", err))
	}

	if err := client.Auth(s.config.Username, s.config.Password); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to authenticate to POP3 server: %v", err))
	}

	err = client.Quit()
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to close POP3 connection: %v", err))
	}

	return nil
}

func (s *Pop3Service) GetEvents(ctx context.Context, regEvents *v1beta1.EventRegistry, lastEventId int, limit int) (events []*v1beta1.EventWithPayload, eventId int, err error) {
	registeredEvent, err := regEvents.Find("EmailReceived")
	if err != nil {
		return nil, lastEventId, fmt.Errorf("failed to find registered event %s: %w", "EmailReceived", err)
	}

	logger := log.FromCtx(ctx)
	logger = logger.With(slog.String("component", "Pop3Service"), slog.String("func", "GetEvents"), slog.Int("lastEventId", lastEventId), slog.Int("limit", limit))

	c, err := s.client.NewConn()

	if err != nil {
		return nil, lastEventId, fmt.Errorf("failed to connect to POP3 server: %v", err)
	}
	defer func() {
		if err := c.Quit(); err != nil {
			logger.Error("failed to quit POP3 connection", log.Err(err))
		}
	}()

	if err := c.Auth(s.config.Username, s.config.Password); err != nil {
		return nil, lastEventId, fmt.Errorf("failed to authenticate to POP3 server: %v", err)
	}

	// Print the total number of messages and their size.
	count, _, err := c.Stat()
	if err != nil {
		return nil, lastEventId, fmt.Errorf("failed to get message count: %v", err)
	}

	if lastEventId < 0 {
		return nil, lastEventId, fmt.Errorf("invalid lastEventId: %d", lastEventId)
	}

	startEventId := lastEventId + 1

	if startEventId > count {
		return events, lastEventId, nil
	}

	maxEventId := max(startEventId+limit, count)

	events = make([]*v1beta1.EventWithPayload, 0, count-startEventId)

	// Pull all messages on the server. Message IDs go from 1 to count.
	for eventId = startEventId; eventId <= maxEventId; eventId++ {
		msg, err := c.Retr(eventId)
		if err != nil {
			fmt.Println("error retrieving message:", err)
			continue
		}

		mr := mail.NewReader(msg)

		var textBody []byte
		var htmlBody []byte

		if mr != nil {
			// This is a multipart message
			for {
				p, err := mr.NextPart()

				if errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					logger.Error("failed to read next part", log.Err(err))
				}

				if textBody == nil {
					logger.Debug("This is a text/plain part")
					textBody, err = io.ReadAll(p.Body)
				} else if htmlBody == nil {
					logger.Debug("This is a text/html part")
					htmlBody, err = io.ReadAll(p.Body)
				} else {
					logger.Debug("This is an unknown part")
					_, err = io.ReadAll(p.Body)
				}

				if err != nil {
					logger.Error("failed to read body", log.Err(err))
				}
			}
		} else {
			logger.Debug("This is a single part message")
			textBody, err = io.ReadAll(msg.Body)
			if err != nil {
				logger.Error("failed to read body", log.Err(err))
			}
		}

		if htmlBody == nil {
			htmlBody = textBody
		}

		layout := "Mon, 02 Jan 2006 15:04:05 -0700"
		dateTime, err := time.Parse(layout, msg.Header.Get("date"))
		if err != nil {
			logger.Error("failed to parse date", log.Err(err))
		}

		from, err := parseAddress(msg.Header.Get("from"))
		if err != nil {
			logger.Error("failed to parse from address", log.Err(err))
		}

		returnPath, err := parseAddress(msg.Header.Get("return-path"))
		if err != nil {
			logger.Error("failed to parse return-path address", log.Err(err))
		}

		to, err := parseAddressList(msg.Header.Get("to"))
		if err != nil {
			logger.Error("failed to parse to address", log.Err(err))
		}

		cc, err := parseAddressList(msg.Header.Get("cc"))
		if err != nil {
			logger.Error("failed to parse cc address", log.Err(err))
		}

		bcc, err := parseAddressList(msg.Header.Get("bcc"))
		if err != nil {
			logger.Error("failed to parse bcc address", log.Err(err))
		}

		event, err := registeredEvent.WithPayload(&emailv1beta1.ReceivedEmail{
			Id: msg.Header.Get("message-id"),
			At: dateTime.Format(time.RFC3339),
			Email: &emailv1beta1.Email{
				From:       from,
				ReturnPath: returnPath,
				To:         to,
				Cc:         cc,
				Bcc:        bcc,
				Subject:    msg.Header.Get("subject"),
				TextBody:   string(textBody),
				HtmlBody:   string(htmlBody),
			},
		})
		if err != nil {
			logger.Error("failed to marshal event payload", log.Err(err))
			continue
		}

		events = append(events, event)
	}

	return events, eventId, nil
}
