package pkg

import (
	"context"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/fivetran"
	fb "github.com/huandu/facebook/v2"
)

type (
	FacebookAdsConfig struct {
		// Number of months' worth of reporting data you'd like to include in your initial sync. This cannot be modified once the connector is created. Default value: `THREE`.
		TimeframeMonths string `json:"timeframe_months" jsonschema:"enum=THREE,enum=SIX,enum=TWELVE,enum=TWENTY_FOUR,enum=ALL_TIME"`
		// Option to select connector should sync all accounts or specific accounts. [Possible sync_mode values](/docs/connectors/applications/facebook-ads-insights/api-config#syncmode).
		SyncMode string `json:"sync_mode" jsonschema:"enum=AllAccounts,enum=SpecificAccounts"`
		// List of accounts of which connector will sync the data.
		Accounts []string `json:"accounts,omitempty"`
	}
)

func NewFacebookAdsSource(config *Config, app *fb.App) fivetran.SourceConfigFunc {
	return func(ctx context.Context) (fivetran.SourceConfig, error) {
		// srcConfig, err := fivetran.NewSourceConfig[*FacebookAdsConfig]("facebook_ads")
		// if err != nil {
		// 	return nil, err
		// }

		// var authConfig = NewAuthConfig(config, []string{
		// 	"business_management",
		// 	"ads_management",
		// 	"ads_read",
		// })

		// srcConfig.SetAuthConfig(authConfig)

		// srcConfig.SetRefreshToken(func(ctx context.Context) (*oauth2.Token, error) {
		// 	token, err := middleware.OAuthTokenFromContext(ctx)
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	return RefreshToken(ctx, app, middleware.OAuthTokenToProto(token))
		// })

		// srcConfig.SetConfigData(func(ctx context.Context) (*structpb.Struct, error) {
		// 	session, err := NewSession(ctx, app, authConfig)
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	resp, err := session.Get("/me/adaccounts", fb.Params{"fields": "business"})
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	var data = map[string]any{
		// 		"accounts": resp["data"],
		// 	}

		// 	return structpb.NewStruct(data)
		// })

		// srcConfig.SetConnectorClientAccess(func(_ context.Context, ca *connectors.ConnectionAuthClientAccess) error {
		// 	ca.ClientID(authConfig.ClientId).ClientSecret(authConfig.ClientSecret)
		// 	return nil
		// })
		// srcConfig.SetConnectorAuth(func(ctx context.Context, ca *connectors.ConnectionAuth) error {

		// 	token, err := middleware.OAuthTokenFromContext(ctx)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	ca.AccessToken(token.AccessToken)
		// 	return nil
		// })

		// return srcConfig, nil
		return nil, nil
	}
}
