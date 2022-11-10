package provisioner

import (
	"context"
	"fmt"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/cert-manager/cert-manager/pkg/util/pki"
	"github.com/go-logr/logr"
	"github.com/oracle/oci-go-sdk/v65/certificatesmanagement"
	"github.com/oracle/oci-go-sdk/v65/common"
	ocicav1alpha1 "github.com/william20111/oci-privateca-issuer/pkg/api/v1alpha1"
	"time"
)

const (
	// DefaultDurationInterval The default validity duration, if not provided.
	DefaultDurationInterval = time.Hour * 24 * 7
	// OCICertManagerTagKey The default tag key on a certificate
	OCICertManagerTagKey = "cert-manager"
	// OCICertManagerTagValue The default tag value on a certificate
	OCICertManagerTagValue = "true"
)

// GenericProvisioner abstracts over the Provisioner type for mocking purposes
type GenericProvisioner interface {
	Sign(ctx context.Context, cr *cmapi.CertificateRequest, log logr.Logger) ([]byte, []byte, error)
}

type ociCAClient interface {
	CreateCertificate(ctx context.Context, request certificatesmanagement.CreateCertificateRequest) (response certificatesmanagement.CreateCertificateResponse, err error)
	GetCertificateAuthority(ctx context.Context, request certificatesmanagement.GetCertificateAuthorityRequest) (response certificatesmanagement.GetCertificateAuthorityResponse, err error)
}

type Provisioner struct {
	ociClient ociCAClient
	logger    logr.Logger
	iss       ocicav1alpha1.OCICAClusterIssuer
}

func New(logger logr.Logger, iss ocicav1alpha1.OCICAClusterIssuer) (*Provisioner, error) {
	certClient, err := certificatesmanagement.NewCertificatesManagementClientWithConfigurationProvider(common.NewRawConfigurationProvider(
		iss.Spec.SecretRef.Tenancy, iss.Spec.SecretRef.User, iss.Spec.SecretRef.Region, iss.Spec.SecretRef.FingerPrint, iss.Spec.SecretRef.PrivateKey, &iss.Spec.SecretRef.PrivateKeyPassphrase,
	))
	if err != nil {
		return nil, err
	}
	p := &Provisioner{
		logger:    logger,
		ociClient: certClient,
		iss:       iss,
	}
	return p, nil
}
func (p *Provisioner) Sign(ctx context.Context, cr *cmapi.CertificateRequest, log logr.Logger) ([]byte, []byte, error) {
	_, err := pki.DecodeX509CertificateRequestBytes(cr.Spec.Request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode CSR for signing: %s", err)
	}
	var expiry time.Time
	start := time.Now().UTC()
	if cr.Spec.Duration == nil {
		expiry = start.Add(DefaultDurationInterval)
	} else {
		expiry = start.Add(cr.Spec.Duration.Duration)
	}

	_, err = p.ociClient.CreateCertificate(ctx, certificatesmanagement.CreateCertificateRequest{
		CreateCertificateDetails: certificatesmanagement.CreateCertificateDetails{
			Name:          common.String(cr.Name),
			CompartmentId: nil,
			CertificateConfig: certificatesmanagement.CreateCertificateManagedExternallyIssuedByInternalCaConfigDetails{
				IssuerCertificateAuthorityId: common.String(p.iss.Spec.OCID),
				CsrPem:                       common.String(string(cr.Spec.Request)),
				VersionName:                  common.String(cr.Name),
				Validity: &certificatesmanagement.Validity{
					TimeOfValidityNotAfter:  &common.SDKTime{Time: expiry},
					TimeOfValidityNotBefore: &common.SDKTime{Time: start},
				},
			},
			Description: common.String(cr.Name),
			FreeformTags: map[string]string{
				OCICertManagerTagKey: OCICertManagerTagValue,
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}
