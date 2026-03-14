package accountoapi

import accountoapigen "github.com/datakit-dev/dtkt-integrations/postmark/pkg/accountoapi/gen"

// These are the generated params with XPostmarkServerToken manually removed with multi cursor edit.

type DeleteDomainParams struct {
	Domainid int
}

type DeleteSenderSignatureParams struct {
	Signatureid int
}

type DeleteServerParams struct {
	Serverid int
}

type EditDomainParams struct {
	Domainid int
}

type EditSenderSignatureParams struct {
	Signatureid int
}

type EditServerInformationParams struct {
	Serverid int
}

type GetDomainParams struct {
	Domainid int
}

type GetSenderSignatureParams struct {
	Signatureid int
}

type GetServerInformationParams struct {
	Serverid int
}

type ListDomainsParams struct {
	Count  int
	Offset int
}

type ListSenderSignaturesParams struct {
	Count  int
	Offset int
}

type ListServersParams struct {
	Count  int
	Offset int
	Name   accountoapigen.OptString
}

type RequestDkimVerificationForDomainParams struct {
	Domainid int
}

type RequestNewDKIMKeyForSenderSignatureParams struct {
	Signatureid int
}

type RequestReturnPathVerificationForDomainParams struct {
	Domainid int
}

type RequestSPFVerificationForDomainParams struct {
	Domainid int
}

type RequestSPFVerificationForSenderSignatureParams struct {
	Signatureid int
}

type ResendSenderSignatureConfirmationEmailParams struct {
	Signatureid int
}

type RotateDKIMKeyForDomainParams struct {
	Domainid int
}

func (p DeleteDomainParams) WithApiToken(apiToken string) accountoapigen.DeleteDomainParams {
	return accountoapigen.DeleteDomainParams{
		XPostmarkAccountToken: apiToken,
		Domainid:              p.Domainid,
	}
}

func (p DeleteSenderSignatureParams) WithApiToken(apiToken string) accountoapigen.DeleteSenderSignatureParams {
	return accountoapigen.DeleteSenderSignatureParams{
		XPostmarkAccountToken: apiToken,
		Signatureid:           p.Signatureid,
	}
}

func (p DeleteServerParams) WithApiToken(apiToken string) accountoapigen.DeleteServerParams {
	return accountoapigen.DeleteServerParams{
		XPostmarkAccountToken: apiToken,
		Serverid:              p.Serverid,
	}
}

func (p EditDomainParams) WithApiToken(apiToken string) accountoapigen.EditDomainParams {
	return accountoapigen.EditDomainParams{
		XPostmarkAccountToken: apiToken,
		Domainid:              p.Domainid,
	}
}

func (p EditSenderSignatureParams) WithApiToken(apiToken string) accountoapigen.EditSenderSignatureParams {
	return accountoapigen.EditSenderSignatureParams{
		XPostmarkAccountToken: apiToken,
		Signatureid:           p.Signatureid,
	}
}

func (p EditServerInformationParams) WithApiToken(apiToken string) accountoapigen.EditServerInformationParams {
	return accountoapigen.EditServerInformationParams{
		XPostmarkAccountToken: apiToken,
		Serverid:              p.Serverid,
	}
}

func (p GetDomainParams) WithApiToken(apiToken string) accountoapigen.GetDomainParams {
	return accountoapigen.GetDomainParams{
		XPostmarkAccountToken: apiToken,
		Domainid:              p.Domainid,
	}
}

func (p GetSenderSignatureParams) WithApiToken(apiToken string) accountoapigen.GetSenderSignatureParams {
	return accountoapigen.GetSenderSignatureParams{
		XPostmarkAccountToken: apiToken,
		Signatureid:           p.Signatureid,
	}
}

func (p GetServerInformationParams) WithApiToken(apiToken string) accountoapigen.GetServerInformationParams {
	return accountoapigen.GetServerInformationParams{
		XPostmarkAccountToken: apiToken,
		Serverid:              p.Serverid,
	}
}

func (p ListDomainsParams) WithApiToken(apiToken string) accountoapigen.ListDomainsParams {
	return accountoapigen.ListDomainsParams{
		XPostmarkAccountToken: apiToken,
		Count:                 p.Count,
		Offset:                p.Offset,
	}
}

func (p ListSenderSignaturesParams) WithApiToken(apiToken string) accountoapigen.ListSenderSignaturesParams {
	return accountoapigen.ListSenderSignaturesParams{
		XPostmarkAccountToken: apiToken,
		Count:                 p.Count,
		Offset:                p.Offset,
	}
}

func (p ListServersParams) WithApiToken(apiToken string) accountoapigen.ListServersParams {
	return accountoapigen.ListServersParams{
		XPostmarkAccountToken: apiToken,
		Count:                 p.Count,
		Offset:                p.Offset,
		Name:                  p.Name,
	}
}

func (p RequestDkimVerificationForDomainParams) WithApiToken(apiToken string) accountoapigen.RequestDkimVerificationForDomainParams {
	return accountoapigen.RequestDkimVerificationForDomainParams{
		XPostmarkAccountToken: apiToken,
		Domainid:              p.Domainid,
	}
}

func (p RequestNewDKIMKeyForSenderSignatureParams) WithApiToken(apiToken string) accountoapigen.RequestNewDKIMKeyForSenderSignatureParams {
	return accountoapigen.RequestNewDKIMKeyForSenderSignatureParams{
		XPostmarkAccountToken: apiToken,
		Signatureid:           p.Signatureid,
	}
}

func (p RequestReturnPathVerificationForDomainParams) WithApiToken(apiToken string) accountoapigen.RequestReturnPathVerificationForDomainParams {
	return accountoapigen.RequestReturnPathVerificationForDomainParams{
		XPostmarkAccountToken: apiToken,
		Domainid:              p.Domainid,
	}
}

func (p RequestSPFVerificationForDomainParams) WithApiToken(apiToken string) accountoapigen.RequestSPFVerificationForDomainParams {
	return accountoapigen.RequestSPFVerificationForDomainParams{
		XPostmarkAccountToken: apiToken,
		Domainid:              p.Domainid,
	}
}

func (p RequestSPFVerificationForSenderSignatureParams) WithApiToken(apiToken string) accountoapigen.RequestSPFVerificationForSenderSignatureParams {
	return accountoapigen.RequestSPFVerificationForSenderSignatureParams{
		XPostmarkAccountToken: apiToken,
		Signatureid:           p.Signatureid,
	}
}

func (p ResendSenderSignatureConfirmationEmailParams) WithApiToken(apiToken string) accountoapigen.ResendSenderSignatureConfirmationEmailParams {
	return accountoapigen.ResendSenderSignatureConfirmationEmailParams{
		XPostmarkAccountToken: apiToken,
		Signatureid:           p.Signatureid,
	}
}

func (p RotateDKIMKeyForDomainParams) WithApiToken(apiToken string) accountoapigen.RotateDKIMKeyForDomainParams {
	return accountoapigen.RotateDKIMKeyForDomainParams{
		XPostmarkAccountToken: apiToken,
		Domainid:              p.Domainid,
	}
}
