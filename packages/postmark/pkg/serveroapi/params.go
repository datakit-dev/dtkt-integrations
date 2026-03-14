package serveroapi

import serveroapigen "github.com/datakit-dev/dtkt-integrations/postmark/pkg/serveroapi/gen"

// These are the generated params with XPostmarkServerToken manually removed with multi cursor edit.

type ActivateBounceParams struct {
	Bounceid int64
}

type BypassRulesForInboundMessageParams struct {
	Messageid string
}

type DeleteInboundRuleParams struct {
	Triggerid int
}

type DeleteTemplateParams struct {
	TemplateIdOrAlias string
}

type GetBounceCountsParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetBounceDumpParams struct {
	Bounceid int64
}

type GetBouncesParams struct {
	Count       int
	Offset      int
	Type        serveroapigen.OptGetBouncesType
	Inactive    serveroapigen.OptBool
	EmailFilter serveroapigen.OptString
	MessageID   serveroapigen.OptString
	Tag         serveroapigen.OptString
	Todate      serveroapigen.OptDate
	Fromdate    serveroapigen.OptDate
}

type GetClicksForSingleOutboundMessageParams struct {
	Messageid string
	Count     int
	Offset    int
}

type GetInboundMessageDetailsParams struct {
	Messageid string
}

type GetOpensForSingleOutboundMessageParams struct {
	Messageid string
	Count     int
	Offset    int
}

type GetOutboundClickCountsParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetOutboundClickCountsByBrowserFamilyParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetOutboundClickCountsByLocationParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetOutboundClickCountsByPlatformParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetOutboundMessageDetailsParams struct {
	Messageid string
}

type GetOutboundMessageDumpParams struct {
	Messageid string
}

type GetOutboundOpenCountsParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetOutboundOpenCountsByEmailClientParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetOutboundOpenCountsByPlatformParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetOutboundOverviewStatisticsParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetSentCountsParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetSingleBounceParams struct {
	Bounceid int64
}

type GetSingleTemplateParams struct {
	TemplateIdOrAlias string
}

type GetSpamComplaintsParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type GetTrackedEmailCountsParams struct {
	Tag      serveroapigen.OptString
	Fromdate serveroapigen.OptDate
	Todate   serveroapigen.OptDate
}

type ListInboundRulesParams struct {
	Count  int
	Offset int
}

type ListTemplatesParams struct {
	Count  int
	Offset int
}

type RetryInboundMessageProcessingParams struct {
	Messageid string
}

type SearchClicksForOutboundMessagesParams struct {
	Count         int
	Offset        int
	Recipient     serveroapigen.OptString
	Tag           serveroapigen.OptString
	ClientName    serveroapigen.OptString
	ClientCompany serveroapigen.OptString
	ClientFamily  serveroapigen.OptString
	OsName        serveroapigen.OptString
	OsFamily      serveroapigen.OptString
	OsCompany     serveroapigen.OptString
	Platform      serveroapigen.OptString
	Country       serveroapigen.OptString
	Region        serveroapigen.OptString
	City          serveroapigen.OptString
}

type SearchInboundMessagesParams struct {
	Count       int
	Offset      int
	Recipient   serveroapigen.OptString
	Fromemail   serveroapigen.OptString
	Subject     serveroapigen.OptString
	Mailboxhash serveroapigen.OptString
	Tag         serveroapigen.OptString
	Status      serveroapigen.OptSearchInboundMessagesStatus
	Todate      serveroapigen.OptDate
	Fromdate    serveroapigen.OptDate
}

type SearchOpensForOutboundMessagesParams struct {
	Count         int
	Offset        int
	Recipient     serveroapigen.OptString
	Tag           serveroapigen.OptString
	ClientName    serveroapigen.OptString
	ClientCompany serveroapigen.OptString
	ClientFamily  serveroapigen.OptString
	OsName        serveroapigen.OptString
	OsFamily      serveroapigen.OptString
	OsCompany     serveroapigen.OptString
	Platform      serveroapigen.OptString
	Country       serveroapigen.OptString
	Region        serveroapigen.OptString
	City          serveroapigen.OptString
}

type SearchOutboundMessagesParams struct {
	Count     int
	Offset    int
	Recipient serveroapigen.OptString
	Fromemail serveroapigen.OptString
	Tag       serveroapigen.OptString
	Status    serveroapigen.OptSearchOutboundMessagesStatus
	Todate    serveroapigen.OptDate
	Fromdate  serveroapigen.OptDate
}

type UpdateTemplateParams struct {
	TemplateIdOrAlias string
}

func (p ActivateBounceParams) WithApiToken(apiToken string) serveroapigen.ActivateBounceParams {
	return serveroapigen.ActivateBounceParams{
		XPostmarkServerToken: apiToken,
		Bounceid:             p.Bounceid,
	}
}

func (p BypassRulesForInboundMessageParams) WithApiToken(apiToken string) serveroapigen.BypassRulesForInboundMessageParams {
	return serveroapigen.BypassRulesForInboundMessageParams{
		XPostmarkServerToken: apiToken,
		Messageid:            p.Messageid,
	}
}

func (p DeleteInboundRuleParams) WithApiToken(apiToken string) serveroapigen.DeleteInboundRuleParams {
	return serveroapigen.DeleteInboundRuleParams{
		XPostmarkServerToken: apiToken,
		Triggerid:            p.Triggerid,
	}
}

func (p DeleteTemplateParams) WithApiToken(apiToken string) serveroapigen.DeleteTemplateParams {
	return serveroapigen.DeleteTemplateParams{
		XPostmarkServerToken: apiToken,
		TemplateIdOrAlias:    p.TemplateIdOrAlias,
	}
}

func (p GetBounceCountsParams) WithApiToken(apiToken string) serveroapigen.GetBounceCountsParams {
	return serveroapigen.GetBounceCountsParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetBounceDumpParams) WithApiToken(apiToken string) serveroapigen.GetBounceDumpParams {
	return serveroapigen.GetBounceDumpParams{
		XPostmarkServerToken: apiToken,
		Bounceid:             p.Bounceid,
	}
}

func (p GetBouncesParams) WithApiToken(apiToken string) serveroapigen.GetBouncesParams {
	return serveroapigen.GetBouncesParams{
		XPostmarkServerToken: apiToken,
		Count:                p.Count,
		Offset:               p.Offset,
		Type:                 p.Type,
		Inactive:             p.Inactive,
		EmailFilter:          p.EmailFilter,
		MessageID:            p.MessageID,
		Tag:                  p.Tag,
		Todate:               p.Todate,
		Fromdate:             p.Fromdate,
	}
}

func (p GetClicksForSingleOutboundMessageParams) WithApiToken(apiToken string) serveroapigen.GetClicksForSingleOutboundMessageParams {
	return serveroapigen.GetClicksForSingleOutboundMessageParams{
		XPostmarkServerToken: apiToken,
		Messageid:            p.Messageid,
		Count:                p.Count,
		Offset:               p.Offset,
	}
}

func (p GetInboundMessageDetailsParams) WithApiToken(apiToken string) serveroapigen.GetInboundMessageDetailsParams {
	return serveroapigen.GetInboundMessageDetailsParams{
		XPostmarkServerToken: apiToken,
		Messageid:            p.Messageid,
	}
}

func (p GetOpensForSingleOutboundMessageParams) WithApiToken(apiToken string) serveroapigen.GetOpensForSingleOutboundMessageParams {
	return serveroapigen.GetOpensForSingleOutboundMessageParams{
		XPostmarkServerToken: apiToken,
		Messageid:            p.Messageid,
		Count:                p.Count,
		Offset:               p.Offset,
	}
}

func (p GetOutboundClickCountsParams) WithApiToken(apiToken string) serveroapigen.GetOutboundClickCountsParams {
	return serveroapigen.GetOutboundClickCountsParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetOutboundClickCountsByBrowserFamilyParams) WithApiToken(apiToken string) serveroapigen.GetOutboundClickCountsByBrowserFamilyParams {
	return serveroapigen.GetOutboundClickCountsByBrowserFamilyParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetOutboundClickCountsByLocationParams) WithApiToken(apiToken string) serveroapigen.GetOutboundClickCountsByLocationParams {
	return serveroapigen.GetOutboundClickCountsByLocationParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetOutboundClickCountsByPlatformParams) WithApiToken(apiToken string) serveroapigen.GetOutboundClickCountsByPlatformParams {
	return serveroapigen.GetOutboundClickCountsByPlatformParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetOutboundMessageDetailsParams) WithApiToken(apiToken string) serveroapigen.GetOutboundMessageDetailsParams {
	return serveroapigen.GetOutboundMessageDetailsParams{
		XPostmarkServerToken: apiToken,
		Messageid:            p.Messageid,
	}
}

func (p GetOutboundMessageDumpParams) WithApiToken(apiToken string) serveroapigen.GetOutboundMessageDumpParams {
	return serveroapigen.GetOutboundMessageDumpParams{
		XPostmarkServerToken: apiToken,
		Messageid:            p.Messageid,
	}
}

func (p GetOutboundOpenCountsParams) WithApiToken(apiToken string) serveroapigen.GetOutboundOpenCountsParams {
	return serveroapigen.GetOutboundOpenCountsParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetOutboundOpenCountsByEmailClientParams) WithApiToken(apiToken string) serveroapigen.GetOutboundOpenCountsByEmailClientParams {
	return serveroapigen.GetOutboundOpenCountsByEmailClientParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetOutboundOpenCountsByPlatformParams) WithApiToken(apiToken string) serveroapigen.GetOutboundOpenCountsByPlatformParams {
	return serveroapigen.GetOutboundOpenCountsByPlatformParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetOutboundOverviewStatisticsParams) WithApiToken(apiToken string) serveroapigen.GetOutboundOverviewStatisticsParams {
	return serveroapigen.GetOutboundOverviewStatisticsParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetSentCountsParams) WithApiToken(apiToken string) serveroapigen.GetSentCountsParams {
	return serveroapigen.GetSentCountsParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetSingleBounceParams) WithApiToken(apiToken string) serveroapigen.GetSingleBounceParams {
	return serveroapigen.GetSingleBounceParams{
		XPostmarkServerToken: apiToken,
		Bounceid:             p.Bounceid,
	}
}

func (p GetSingleTemplateParams) WithApiToken(apiToken string) serveroapigen.GetSingleTemplateParams {
	return serveroapigen.GetSingleTemplateParams{
		XPostmarkServerToken: apiToken,
		TemplateIdOrAlias:    p.TemplateIdOrAlias,
	}
}

func (p GetSpamComplaintsParams) WithApiToken(apiToken string) serveroapigen.GetSpamComplaintsParams {
	return serveroapigen.GetSpamComplaintsParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p GetTrackedEmailCountsParams) WithApiToken(apiToken string) serveroapigen.GetTrackedEmailCountsParams {
	return serveroapigen.GetTrackedEmailCountsParams{
		XPostmarkServerToken: apiToken,
		Tag:                  p.Tag,
		Fromdate:             p.Fromdate,
		Todate:               p.Todate,
	}
}

func (p ListInboundRulesParams) WithApiToken(apiToken string) serveroapigen.ListInboundRulesParams {
	return serveroapigen.ListInboundRulesParams{
		XPostmarkServerToken: apiToken,
		Count:                p.Count,
		Offset:               p.Offset,
	}
}

func (p ListTemplatesParams) WithApiToken(apiToken string) serveroapigen.ListTemplatesParams {
	return serveroapigen.ListTemplatesParams{
		XPostmarkServerToken: apiToken,
		Count:                p.Count,
		Offset:               p.Offset,
	}
}

func (p RetryInboundMessageProcessingParams) WithApiToken(apiToken string) serveroapigen.RetryInboundMessageProcessingParams {
	return serveroapigen.RetryInboundMessageProcessingParams{
		XPostmarkServerToken: apiToken,
		Messageid:            p.Messageid,
	}
}

func (p SearchClicksForOutboundMessagesParams) WithApiToken(apiToken string) serveroapigen.SearchClicksForOutboundMessagesParams {
	return serveroapigen.SearchClicksForOutboundMessagesParams{
		XPostmarkServerToken: apiToken,
		Count:                p.Count,
		Offset:               p.Offset,
		Recipient:            p.Recipient,
		Tag:                  p.Tag,
		ClientName:           p.ClientName,
		ClientCompany:        p.ClientCompany,
		ClientFamily:         p.ClientFamily,
		OsName:               p.OsName,
		OsFamily:             p.OsFamily,
		OsCompany:            p.OsCompany,
		Platform:             p.Platform,
		Country:              p.Country,
		Region:               p.Region,
		City:                 p.City,
	}
}

func (p SearchInboundMessagesParams) WithApiToken(apiToken string) serveroapigen.SearchInboundMessagesParams {
	return serveroapigen.SearchInboundMessagesParams{
		XPostmarkServerToken: apiToken,
		Count:                p.Count,
		Offset:               p.Offset,
		Recipient:            p.Recipient,
		Fromemail:            p.Fromemail,
		Subject:              p.Subject,
		Mailboxhash:          p.Mailboxhash,
		Tag:                  p.Tag,
		Status:               p.Status,
		Todate:               p.Todate,
		Fromdate:             p.Fromdate,
	}
}

func (p SearchOpensForOutboundMessagesParams) WithApiToken(apiToken string) serveroapigen.SearchOpensForOutboundMessagesParams {
	return serveroapigen.SearchOpensForOutboundMessagesParams{
		XPostmarkServerToken: apiToken,
		Count:                p.Count,
		Offset:               p.Offset,
		Recipient:            p.Recipient,
		Tag:                  p.Tag,
		ClientName:           p.ClientName,
		ClientCompany:        p.ClientCompany,
		ClientFamily:         p.ClientFamily,
		OsName:               p.OsName,
		OsFamily:             p.OsFamily,
		OsCompany:            p.OsCompany,
		Platform:             p.Platform,
		Country:              p.Country,
		Region:               p.Region,
		City:                 p.City,
	}
}

func (p SearchOutboundMessagesParams) WithApiToken(apiToken string) serveroapigen.SearchOutboundMessagesParams {
	return serveroapigen.SearchOutboundMessagesParams{
		XPostmarkServerToken: apiToken,
		Count:                p.Count,
		Offset:               p.Offset,
		Recipient:            p.Recipient,
		Fromemail:            p.Fromemail,
		Tag:                  p.Tag,
		Status:               p.Status,
		Todate:               p.Todate,
		Fromdate:             p.Fromdate,
	}
}

func (p UpdateTemplateParams) WithApiToken(apiToken string) serveroapigen.UpdateTemplateParams {
	return serveroapigen.UpdateTemplateParams{
		XPostmarkServerToken: apiToken,
		TemplateIdOrAlias:    p.TemplateIdOrAlias,
	}
}
