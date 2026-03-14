package pkg

type (
	FivetranInstagramConfig struct {
		// Number of months' worth of reporting data you'd like to include in your initial sync. This cannot be modified once the connector is created. Default value: `THREE`.
		TimeframeMonths string `json:"timeframe_months" jsonschema:"enum=THREE,enum=SIX,enum=TWELVE,enum=TWENTY_FOUR,enum=ALL_TIME"`
		// Option to select connector should sync all accounts or specific accounts. [Possible sync_mode values](/docs/connectors/applications/facebook-ads-insights/api-config#syncmode).
		SyncMode string `json:"sync_mode" jsonschema:"enum=AllAccounts,enum=SpecificAccounts"`
		// List of accounts of which connector will sync the data.
		Accounts []string `json:"accounts,omitempty"`
	}
	Business struct {
		ID   string `json:"id" facebook:",required"`
		Name string `json:"name" facbook:"name"`
	}
	InstagramAccount struct {
		ID   string `json:"id" facebook:",required"`
		Name string `json:"name" facbook:"username"`
	}
)

// func NewFivetranInstagramSource(app *fb.App, config *Config) *SourceConfig {
// 	var authConfig = NewAuthConfig(config, []string{
// 		"instagram_basic",
// 		"instagram_manage_insights",
// 		"pages_show_list",
// 		"pages_read_engagement",
// 	})

// 	return &SourceConfig{
// 		SourceConfig: fivetran.NewSourceConfigMust[*FivetranAdsConfig]("instagram_business"),
// 		AuthConfig:   authConfig,
// 		ConfigData: func(ctx context.Context) (*structpb.Struct, error) {
// 			session, err := NewSession(ctx, app, authConfig)
// 			if err != nil {
// 				return nil, err
// 			}

// 			resp, err := session.Get("/me/businesses", fb.Params{"fields": "name"})
// 			if err != nil {
// 				return nil, err
// 			}

// 			var accounts []any
// 			if bizs, ok := resp["data"].([]any); ok {
// 				for bizMap := range slices.Values(bizs) {
// 					biz, err := common.FromMap[*Business](bizMap.(map[string]any))
// 					if err != nil {
// 						continue
// 					}

// 					resp, err := session.Get("/"+biz.ID+"/instagram_accounts", fb.Params{"fields": "username"})
// 					if err != nil {
// 						return nil, err
// 					}

// 					if resp, ok := resp["data"].([]any); ok {
// 						accounts = append(accounts, resp)
// 					}
// 				}
// 			}

// 			var data = map[string]any{
// 				"accounts": accounts,
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

// 		// 		config, err := common.Unmarshal[*FivetranInstagramConfig](configStr)
// 		// 		if err != nil {
// 		// 			return nil, err
// 		// 		}

// 		// 		cc.SyncMode(config.SyncMode).TimeframeMonths(config.TimeframeMonths)

// 		// 		if config.SyncMode == "SpecificAccounts" {
// 		// 			cc.Accounts(config.Accounts)
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
