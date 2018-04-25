package client

import (
	pbs "github.com/datacol-io/datacol/api/controller"
)

func (c *Client) CreateCertificate(app, domain, cert, key string) error {
	_, err := c.ProviderServiceClient.CertificateCreate(ctx, &pbs.CertificateReq{
		App:         app,
		Domain:      domain,
		CertEncoded: cert,
		KeyEncoded:  key,
	})

	return err
}

func (c *Client) DeleteCertificate(app, domain string) error {
	_, err := c.ProviderServiceClient.CertificateDelete(ctx, &pbs.CertificateReq{
		App:    app,
		Domain: domain,
	})

	return err
}
