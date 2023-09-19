package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/clock"
	"reflect"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestCertificateRequestReconciler_Reconcile(t *testing.T) {
	type fields struct {
		Log                    logr.Logger
		Scheme                 *runtime.Scheme
		Recorder               record.EventRecorder
		Clock                  clock.Clock
		CheckApprovedCondition bool
	}
	type args struct {
		ctx context.Context
		req controllerruntime.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    controllerruntime.Result
		wantErr bool
		objects []client.Object
	}{
		{
			name: "valid sign",
			fields: fields{
				Log:                    logr.Logger{},
				Scheme:                 runtime.NewScheme(),
				Recorder:               record.NewFakeRecorder(10),
				Clock:                  nil,
				CheckApprovedCondition: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CertificateRequestReconciler{
				Client: fake.NewClientBuilder().
					WithScheme(tt.fields.Scheme).
					WithObjects(tt.objects...).
					Build(),
				Log:                    tt.fields.Log,
				Scheme:                 tt.fields.Scheme,
				Recorder:               tt.fields.Recorder,
				Clock:                  tt.fields.Clock,
				CheckApprovedCondition: tt.fields.CheckApprovedCondition,
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
