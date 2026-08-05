package config

import "testing"

func TestCORS_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cors    CORS
		wantErr bool
	}{
		{
			name:    "wildcard origin without credentials is allowed",
			cors:    CORS{AllowOrigins: []string{"*"}, AllowCredentials: false},
			wantErr: false,
		},
		{
			name:    "explicit origins with credentials is allowed",
			cors:    CORS{AllowOrigins: []string{"https://example.com"}, AllowCredentials: true},
			wantErr: false,
		},
		{
			name:    "wildcard origin with credentials is rejected",
			cors:    CORS{AllowOrigins: []string{"*"}, AllowCredentials: true},
			wantErr: true,
		},
		{
			name:    "wildcard among explicit origins with credentials is rejected",
			cors:    CORS{AllowOrigins: []string{"https://example.com", "*"}, AllowCredentials: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cors.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
