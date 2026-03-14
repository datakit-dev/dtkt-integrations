package pkg

type (
	FivetranPagesConfig struct {
		// Number of months' worth of reporting data you'd like to include in your initial sync. This cannot be modified once the connector is created. Default value: `THREE`.
		TimeframeMonths string `json:"timeframe_months" jsonschema:"enum=THREE,enum=SIX,enum=TWELVE,enum=TWENTY_FOUR,enum=ALL_TIME"`
		// Whether to sync all accounts or specific accounts. Default value: AllPages
		SyncMode string `json:"sync_mode" jsonschema:"enum=AllPages,enum=SpecificPages"`
		// Specific pages to sync. Must be populated if sync_mode is set to SpecificPages.
		Pages []string `json:"pages,omitempty"`
	}
)

// func NewFivetranPagesSource(app *fb.App, config *Config) *SourceConfig {
// 	var authConfig = NewAuthConfig(config, []string{
// 		"instagram_basic",
// 		"instagram_manage_insights",
// 		"pages_show_list",
// 		"pages_read_engagement",
// 	})

// 	return &SourceConfig{
// 		SourceConfig: fivetran.NewSourceConfigMust[*FivetranAdsConfig]("facebook_pages"),
// 		AuthConfig:   authConfig,
// 		ConfigData: func(ctx context.Context) (*structpb.Struct, error) {
// 			session, err := NewSession(ctx, app, authConfig)
// 			if err != nil {
// 				return nil, err
// 			}

// 			resp, err := session.Get("/me/accounts", nil)
// 			if err != nil {
// 				return nil, err
// 			}

// 			var data = map[string]any{
// 				"pages": resp["data"],
// 			}

// 			return structpb.NewStruct(data)
// 		},
// 		// Create: func(ctx context.Context, req *replicationv1beta.CreateSourceRequest) (*replicationv1beta1.CreateSourceResponse, error) {
// 		// 	var (
// 		// 		auth = &connectors.ConnectorAuth{}
// 		// 		ca   = &connectors.ConnectorAuthClientAccess{}
// 		// 		cc   = &connectors.ConnectorConfig{}
// 		// 	)

// 		// 	if req.Config != nil {
// 		// 		configStr, err := common.Marshal[string](req.Config.AsMap())
// 		// 		if err != nil {
// 		// 			return nil, err
// 		// 		}

// 		// 		if _, err := configSchema.Validate(configStr); err != nil {
// 		// 			return nil, err
// 		// 		}

// 		// 		config, err := common.Unmarshal[*FivetranPagesConfig](configStr)
// 		// 		if err != nil {
// 		// 			return nil, err
// 		// 		}

// 		// 		cc.SyncMode(config.SyncMode).TimeframeMonths(config.TimeframeMonths)

// 		// 		if config.SyncMode == "SpecificAccounts" {
// 		// 			cc.Accounts(config.Pages)
// 		// 		}
// 		// 	} else {
// 		// 		return nil, fmt.Errorf("facebook ads requires valid config")
// 		// 	}

// 		// 	token, err := middleware.OAuthTokenFromContext(ctx)
// 		// 	if err != nil {
// 		// 		return nil, err
// 		// 	}

// 		// 	return fivetran.CreateSource(ctx,
// 		// 		req.ServiceConfig.GetFivetranConfig(),
// 		// 		configID,
// 		// 		auth.
// 		// 			AccessToken(token.AccessToken).
// 		// 			ClientAccess(
// 		// 				ca.
// 		// 					ClientID(app.AppId).
// 		// 					ClientSecret(app.AppSecret),
// 		// 			),
// 		// 		cc.Schema(req.Name),
// 		// 	)
// 		// },
// 	}
// }
