package pkg

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	sheetsv1beta "github.com/datakit-dev/dtkt-integrations/sheets/pkg/proto/integration/sheets/v1beta"
)

var defaultScopes = []string{
	drive.DriveMetadataScope,
	sheets.SpreadsheetsScope,
}

// Integration instance struct
type Instance struct {
	config *sheetsv1beta.Config

	drive  *drive.Service
	sheets *sheets.Service

	tokenSource    oauth2.TokenSource
	newTokenSource func(*oauth2.Config, *oauth2.Token) oauth2.TokenSource

	mut sync.Mutex
}

// Creates a new instance
func NewInstance(ctx context.Context, config *sheetsv1beta.Config) (*Instance, error) {
	var (
		inst = &Instance{
			config: config,
		}
		opts []option.ClientOption
	)

	if config.GetCredentialsJson() != nil {
		b, err := protojson.Marshal(config.GetCredentialsJson())
		if err != nil {
			return nil, err
		}

		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, b))
	} else if config.GetOauthConfig() != nil {
		inst.newTokenSource = func(config *oauth2.Config, token *oauth2.Token) oauth2.TokenSource {
			return config.TokenSource(ctx, token)
		}

		opts = append(opts, option.WithTokenSource(oauth2.ReuseTokenSource(nil, inst)))
	} else {
		return nil, fmt.Errorf("config.auth requires one of: credentials json or oauth config")
	}

	opts = append(opts, option.WithLogger(log.FromCtx(ctx)))

	drive, err := drive.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	sheets, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	inst.drive = drive
	inst.sheets = sheets

	return inst, nil
}

func (i *Instance) Token() (*oauth2.Token, error) {
	i.mut.Lock()
	defer i.mut.Unlock()

	if i.config.GetOauthConfig() == nil {
		return nil, status.Error(codes.FailedPrecondition, "oauth not configured")
	} else if i.tokenSource == nil {
		return nil, status.Error(codes.FailedPrecondition, "oauth not completed")
	}

	return i.tokenSource.Token()
}

// Orchestrates OAuth checks/exchanges.
func (i *Instance) CheckAuth(ctx context.Context, req *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	if i.config.GetOauthConfig() == nil {
		return nil, status.Error(codes.FailedPrecondition, "oauth not configured")
	} else if req.Type == nil {
		return &basev1beta1.CheckAuthResponse{
			AuthRequired: sharedv1beta1.AuthType_AUTH_TYPE_OAUTH_CODE,
		}, nil
	}

	config, opts := getOAuthConfig(i.config.GetOauthConfig())

	switch req := req.Type.(type) {
	case *basev1beta1.CheckAuthRequest_CodeRequest:
		if req.CodeRequest.GetRedirectUrl() != "" {
			config.RedirectURL = req.CodeRequest.GetRedirectUrl()
		}

		if req.CodeRequest.GetCodeChallenge() != "" && req.CodeRequest.GetCodeChallengeMethod() == sharedv1beta1.CodeChallengeMethod_CODE_CHALLENGE_METHOD_S256 {
			opts = append(opts, oauth2.S256ChallengeOption(req.CodeRequest.GetCodeChallenge()))
		}

		return &basev1beta1.CheckAuthResponse{
			Type: &basev1beta1.CheckAuthResponse_OauthCodeUrl{
				OauthCodeUrl: config.AuthCodeURL(req.CodeRequest.GetState(), opts...),
			},
		}, nil
	case *basev1beta1.CheckAuthRequest_TokenRequest:
		if req.TokenRequest.GetRedirectUrl() != "" {
			config.RedirectURL = req.TokenRequest.GetRedirectUrl()
		}

		if req.TokenRequest.GetCodeVerifier() != "" {
			opts = append(opts, oauth2.VerifierOption(req.TokenRequest.GetCodeVerifier()))
		}

		token, err := config.Exchange(ctx, req.TokenRequest.GetCode(), opts...)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}

		i.mut.Lock()
		i.tokenSource = i.newTokenSource(config, token)
		i.mut.Unlock()

		return &basev1beta1.CheckAuthResponse{
			Type: &basev1beta1.CheckAuthResponse_Token{
				Token: v1beta1.OAuthTokenToProto(token),
			},
		}, nil
	case *basev1beta1.CheckAuthRequest_RefreshRequest:
		if req.RefreshRequest.Token != nil {
			i.mut.Lock()
			i.tokenSource = i.newTokenSource(config, v1beta1.OAuthTokenFromProto(req.RefreshRequest.Token))
			i.mut.Unlock()
		}

		token, err := i.Token()
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}

		return &basev1beta1.CheckAuthResponse{
			Type: &basev1beta1.CheckAuthResponse_Token{
				Token: v1beta1.OAuthTokenToProto(token),
			},
		}, nil
	}

	return &basev1beta1.CheckAuthResponse{}, nil
}

// Close is called to release instance resources (e.g. client connections)
func (i *Instance) Close() error {
	return nil
}

func getOAuthConfig(config *sharedv1beta1.OAuthConfig) (*oauth2.Config, []oauth2.AuthCodeOption) {
	config = proto.CloneOf(config)
	for _, scope := range defaultScopes {
		if !slices.Contains(config.Scopes, scope) {
			config.Scopes = append(config.Scopes, scope)
		}
	}
	return v1beta1.OAuthConfigFromProto(config)
}
