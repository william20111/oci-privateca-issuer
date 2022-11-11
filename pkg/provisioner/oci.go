package provisioner

import (
	"context"
	"fmt"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/cert-manager/cert-manager/pkg/util/pki"
	"github.com/go-logr/logr"
	"github.com/oracle/oci-go-sdk/v65/certificates"
	"github.com/oracle/oci-go-sdk/v65/certificatesmanagement"
	"github.com/oracle/oci-go-sdk/v65/common"
	ocicav1alpha1 "github.com/william20111/oci-privateca-issuer/pkg/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sync"
	"time"
)

const (
	// DefaultDurationInterval The default validity duration, if not provided.
	DefaultDurationInterval = time.Hour * 24 * 7
	// OCICertManagerTagKey The default tag key on a certificate
	OCICertManagerTagKey = "cert-manager"
	// OCICertManagerTagValue The default tag value on a certificate
	OCICertManagerTagValue          = "true"
	OCICertificatePrivateBundleType = "CERTIFICATE_CONTENT_WITH_PRIVATE_KEY"
)

// GenericProvisioner abstracts over the Provisioner type for mocking purposes
type GenericProvisioner interface {
	Sign(ctx context.Context, cr *cmapi.CertificateRequest, log logr.Logger) ([]byte, []byte, error)
}

type ociCAClient interface {
	CreateCertificate(ctx context.Context, request certificatesmanagement.CreateCertificateRequest) (response certificatesmanagement.CreateCertificateResponse, err error)
	GetCertificateAuthority(ctx context.Context, request certificatesmanagement.GetCertificateAuthorityRequest) (response certificatesmanagement.GetCertificateAuthorityResponse, err error)
}

type ociCertificateClient interface {
	GetCertificateBundle(ctx context.Context, request certificates.GetCertificateBundleRequest) (response certificates.GetCertificateBundleResponse, err error)
}

// Collection stores cached Provisioners, stored by namespaced names of the
// issuer.
type Collection struct {
	m sync.Map
}

// Store adds a provisioner to the collection.
func (c *Collection) Store(namespacedName types.NamespacedName, provisioner *Provisioner) {
	c.m.Store(namespacedName, provisioner)
}

type Provisioner struct {
	caClient          ociCAClient
	certificateClient ociCertificateClient
	logger            logr.Logger
	iss               ocicav1alpha1.OCICAClusterIssuer
}

func New(logger logr.Logger, iss ocicav1alpha1.OCICAClusterIssuer) (*Provisioner, error) {
	caClient, err := certificatesmanagement.NewCertificatesManagementClientWithConfigurationProvider(common.NewRawConfigurationProvider(
		iss.Spec.SecretRef.Tenancy, iss.Spec.SecretRef.User, iss.Spec.SecretRef.Region, iss.Spec.SecretRef.FingerPrint, iss.Spec.SecretRef.PrivateKey, &iss.Spec.SecretRef.PrivateKeyPassphrase,
	))
	certClient, err := certificates.NewCertificatesClientWithConfigurationProvider(common.NewRawConfigurationProvider(
		iss.Spec.SecretRef.Tenancy, iss.Spec.SecretRef.User, iss.Spec.SecretRef.Region, iss.Spec.SecretRef.FingerPrint, iss.Spec.SecretRef.PrivateKey, &iss.Spec.SecretRef.PrivateKeyPassphrase,
	))
	if err != nil {
		return nil, err
	}
	p := &Provisioner{
		logger:            logger,
		caClient:          caClient,
		certificateClient: certClient,
		iss:               iss,
	}
	return p, nil
}

func (p *Provisioner) Validate(ctx context.Context) error {
	res, err := p.caClient.GetCertificateAuthority(ctx, certificatesmanagement.GetCertificateAuthorityRequest{
		CertificateAuthorityId: common.String(p.iss.Spec.OCID),
	})
	if err != nil {
		p.logger.Error(err, "cant get certificate authority")
		return err
	}
	if res.Id != &p.iss.Spec.OCID {
		return fmt.Errorf("cant find the certificate authority")
	}
	return nil
}

func (p *Provisioner) Sign(ctx context.Context, cr *cmapi.CertificateRequest, log logr.Logger) ([]byte, error) {
	_, err := pki.DecodeX509CertificateRequestBytes(cr.Spec.Request)
	if err != nil {
		return nil, fmt.Errorf("failed to decode CSR for signing: %s", err)
	}
	var expiry time.Time
	start := time.Now().UTC()
	if cr.Spec.Duration == nil {
		expiry = start.Add(DefaultDurationInterval)
	} else {
		expiry = start.Add(cr.Spec.Duration.Duration)
	}
	certificateSignResponse, err := p.caClient.CreateCertificate(ctx, certificatesmanagement.CreateCertificateRequest{
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
		return nil, err
	}
	res, err := p.certificateClient.GetCertificateBundle(ctx, certificates.GetCertificateBundleRequest{
		CertificateId:          certificateSignResponse.Id,
		CertificateVersionName: common.String(cr.Name),
		CertificateBundleType:  OCICertificatePrivateBundleType,
	})
	if err != nil {
		p.logger.Error(err, "failed fetching certificate")
		return nil, err
	}

	chainPem := res.GetCertChainPem()
	return []byte(*chainPem), nil
}
