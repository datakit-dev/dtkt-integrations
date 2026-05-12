package chrome

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"maps"
	"os"
	"sync"

	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/encoding"
	dtktlog "github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type (
	ChromeService struct {
		browserv1beta.UnimplementedChromeServiceServer
		log    *slog.Logger
		config *browserv1beta.ChromeConfig

		readChs map[uuid.UUID]chan []byte
		writeCh chan []byte

		grp errgroup.Group
		mut sync.Mutex
	}
	ReaderFunc func() ([]byte, error)
)

func NewChromeService(ctx context.Context, log *slog.Logger, config *browserv1beta.ChromeConfig) *ChromeService {
	svc := &ChromeService{
		log:     log,
		config:  config,
		readChs: map[uuid.UUID]chan []byte{},
		writeCh: make(chan []byte, 1),
	}

	svc.startWriter(ctx)
	svc.startReader(ctx)

	context.AfterFunc(ctx, func() {
		err := svc.Close()
		if err != nil {
			log.Debug("ChromeService close", dtktlog.Err(err))
		}
	})

	return svc
}

func (s *ChromeService) SendChromeAction(ctx context.Context, req *browserv1beta.SendChromeActionRequest) (*browserv1beta.SendChromeActionResponse, error) {
	s.log.Info("SendChromeAction called.")
	defer s.log.Info("SendChromeAction done.")

	data, err := encoding.ToJSONV2(req.GetAction())
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.writeCh <- data:
		return &browserv1beta.SendChromeActionResponse{}, nil
	}
}

func (s *ChromeService) StreamChromeEvents(req *browserv1beta.StreamChromeEventsRequest, stream grpc.ServerStreamingServer[browserv1beta.StreamChromeEventsResponse]) error {
	s.log.Info("StreamChromeEvents starting...")
	defer s.log.Info("StreamChromeEvents done.")

	id := uuid.New()
	reader := s.addReader(stream.Context(), id)
	defer s.removeReader(id)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			data, err := reader()
			if err != nil {
				return err
			}

			event := new(browserv1beta.ChromeEvent)
			err = encoding.FromJSONV2(data, event)
			if err != nil {
				return err
			}

			err = stream.Send(&browserv1beta.StreamChromeEventsResponse{
				Event: event,
			})
			if err != nil {
				return err
			}
		}
	}
}

func (s *ChromeService) addReader(ctx context.Context, id uuid.UUID) ReaderFunc {
	readCh := make(chan []byte, 1)
	reader := func() ([]byte, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case data, ok := <-readCh:
			if !ok {
				return nil, io.EOF
			}

			return data, nil
		}
	}

	s.mut.Lock()
	s.readChs[id] = readCh
	s.mut.Unlock()

	return reader
}

func (s *ChromeService) removeReader(id uuid.UUID) {
	s.mut.Lock()
	if readCh, ok := s.readChs[id]; ok {
		close(readCh)
		delete(s.readChs, id)
	}
	s.mut.Unlock()
}

func (s *ChromeService) startReader(ctx context.Context) {
	s.grp.Go(func() error {
		s.log.Info("Reader starting...")
		defer s.log.Info("Reader done.")

		reader := bufio.NewReader(os.Stdin)
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				_, err := reader.Peek(4)
				if err != nil {
					if errors.Is(err, io.EOF) {
						return err
					} else if !errors.Is(err, bufio.ErrBufferFull) {
						s.log.Error(err.Error())
						continue
					}
				}

				dataLen := make([]byte, 4)
				_, err = io.ReadFull(reader, dataLen)
				if err != nil {
					return err
				}

				data := make([]byte, binary.LittleEndian.Uint32(dataLen))
				_, err = io.ReadFull(reader, data)
				if err != nil {
					return err
				}

				s.mut.Lock()
				for _, readCh := range s.readChs {
					select {
					case <-ctx.Done():
						return nil
					case readCh <- data:
					}
				}
				s.mut.Unlock()
			}
		}
	})
}

func (s *ChromeService) startWriter(ctx context.Context) {
	s.grp.Go(func() error {
		s.log.Info("Writer starting...")
		defer s.log.Info("Writer done.")

		data, err := encoding.ToJSONV2(s.config)
		if err != nil {
			return err
		}

		s.log.Info("Writing chrome config.")

		select {
		case <-ctx.Done():
			return nil
		case s.writeCh <- data:
		}

		writer := bufio.NewWriter(os.Stdout)
		for {
			select {
			case <-ctx.Done():
				return nil
			case data, ok := <-s.writeCh:
				if !ok {
					return nil
				}

				dataLen := make([]byte, 4)
				binary.LittleEndian.PutUint32(dataLen, uint32(len(data)))

				_, err := writer.Write(dataLen)
				if err != nil {
					return err
				}

				_, err = writer.Write(data)
				if err != nil {
					return err
				}

				err = writer.Flush()
				if err != nil {
					return err
				}
			}
		}
	})
}

func (s *ChromeService) Close() error {
	close(s.writeCh)

	s.mut.Lock()
	ids := maps.Keys(s.readChs)
	s.mut.Unlock()

	for id := range ids {
		s.removeReader(id)
	}

	s.log.Info("Chrome server stopping...")

	return errors.Join(
		s.grp.Wait(),
		dtktlog.CloseLogger(s.log),
	)
}
