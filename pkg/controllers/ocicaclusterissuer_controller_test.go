package controllers

import (
	"context"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/william20111/oci-privateca-issuer/pkg/api/v1alpha1"
	"github.com/william20111/oci-privateca-issuer/pkg/provisioner"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_validateIssuer(t *testing.T) {
	type args struct {
		spec v1alpha1.OCICAClusterIssuerSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid validation",
			args: args{spec: v1alpha1.OCICAClusterIssuerSpec{
				TenancyID:     "test",
				CompartmentID: "test",
				AuthorityID:   "test",
			}},
			wantErr: false,
		},
		{
			name: "missing authority ID",
			args: args{spec: v1alpha1.OCICAClusterIssuerSpec{
				TenancyID:     "test",
				CompartmentID: "test",
				AuthorityID:   "",
			}},
			wantErr: true,
		},
		{
			name: "missing compartment ID",
			args: args{spec: v1alpha1.OCICAClusterIssuerSpec{
				TenancyID:     "test",
				CompartmentID: "",
				AuthorityID:   "test",
			}},
			wantErr: true,
		},
		{
			name: "midding tenancy ID",
			args: args{spec: v1alpha1.OCICAClusterIssuerSpec{
				TenancyID:     "",
				CompartmentID: "test",
				AuthorityID:   "test",
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateIssuer(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("validateIssuer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOCICAClusterIssuerReconciler_Reconcile(t *testing.T) {
	type fields struct {
		collection *provisioner.Collection
		Scheme     *runtime.Scheme
	}
	type args struct {
		ctx context.Context
		req controllerruntime.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		objects []client.Object
		want    controllerruntime.Result
		wantErr bool
	}{
		{
			name: "valid sign",
			fields: fields{
				collection: &provisioner.Collection{},
				Scheme:     runtime.NewScheme(),
			},
			args: args{
				ctx: context.TODO(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "ns1",
						Name:      "issuer1",
					},
				},
			},
			objects: []client.Object{
				&v1alpha1.OCICAClusterIssuer{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Issuer",
						APIVersion: "v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "issuer1",
						Namespace: "ns1",
					},
					Spec: v1alpha1.OCICAClusterIssuerSpec{
						TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f44n2b2m2yt2j6rx32uzr4h25vqstifsfdsq",
						CompartmentID: "ocid1.compartment.oc1.phx.aaaaaaaaba3pv6wkcr4jqae5f44n2b2m2yt2j6rx32uzr4h25vqstifsfdsq",
						AuthorityID:   "ocid1.certificateauthority.oc1.phx.aaaaaaaaba3pv6wkcr4jqae5f44n2b2m2yt2j6rx32uzr4h25vqstifsfdsq",
					},
					Status: v1alpha1.OCICAClusterIssuerStatus{
						Conditions: []metav1.Condition{
							{
								Type:   string(v1alpha1.ConditionReady),
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			wantErr: false,
			want:    controllerruntime.Result{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmapi.AddToScheme(tt.fields.Scheme)
			v1.AddToScheme(tt.fields.Scheme)
			v1alpha1.AddToScheme(tt.fields.Scheme)
			r := &OCICAClusterIssuerReconciler{
				collection: tt.fields.collection,
				Client: fake.NewClientBuilder().
					WithScheme(tt.fields.Scheme).
					WithObjects(tt.objects...).
					Build(),
				Scheme: tt.fields.Scheme,
			}
			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reconcile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
