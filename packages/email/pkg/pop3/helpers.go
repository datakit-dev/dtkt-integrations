package pop3

import (
	"net/mail"

	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"
)

func parseAddress(address string) (*emailv1beta1.EmailAddress, error) {
	addr, err := mail.ParseAddress(address)
	if err != nil {
		return nil, err
	}

	return &emailv1beta1.EmailAddress{
		Name:    addr.Name,
		Address: addr.Address,
	}, nil
}

func parseAddressList(list string) ([]*emailv1beta1.EmailAddress, error) {
	addrs, err := mail.ParseAddressList(list)
	if err != nil {
		return nil, err
	}

	emailAddrs := make([]*emailv1beta1.EmailAddress, len(addrs))
	for i, addr := range addrs {
		emailAddrs[i] = &emailv1beta1.EmailAddress{
			Name:    addr.Name,
			Address: addr.Address,
		}
	}

	return emailAddrs, nil
}
