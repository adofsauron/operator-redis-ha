package k8sutils

import "testing"

func TestCheckStatefulSetExist(t *testing.T) {
	type args struct {
		namespace string
		stateful  string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckStatefulSetExist(tt.args.namespace, tt.args.stateful)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckStatefulSetExist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckStatefulSetExist() = %v, want %v", got, tt.want)
			}
		})
	}
}
