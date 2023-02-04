package kpture

import "testing"

func TestInvalidEnvParamError_Error(t *testing.T) {
	type fields struct {
		param string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test error",
			fields: fields{
				param: "test",
			},
			want: "invalid environment parameter: test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := InvalidEnvParamError{
				param: tt.fields.param,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("InvalidEnvParamError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
