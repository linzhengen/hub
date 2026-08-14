package config

import "testing"

func TestAnthropic_Validate(t *testing.T) {
	tests := []struct {
		name      string
		anthropic Anthropic
		wantErr   bool
	}{
		{
			name:      "the real client needs an api key",
			anthropic: Anthropic{Mock: false, APIKey: ""},
			wantErr:   true,
		},
		{
			name:      "the real client with an api key is allowed",
			anthropic: Anthropic{Mock: false, APIKey: "sk-ant-test"},
			wantErr:   false,
		},
		{
			name:      "the mock backend needs no api key",
			anthropic: Anthropic{Mock: true, APIKey: ""},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.anthropic.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// EnvConfig.Validate must run every section: a check nobody calls is a check
// that does not exist.
func TestEnvConfig_ValidateRunsEverySection(t *testing.T) {
	valid := EnvConfig{
		CORS:      CORS{AllowOrigins: []string{"*"}},
		Anthropic: Anthropic{Mock: true},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}

	badCORS := valid
	badCORS.CORS.AllowCredentials = true
	if err := badCORS.Validate(); err == nil {
		t.Error("Validate() = nil, want the CORS error")
	}

	badAnthropic := valid
	badAnthropic.Anthropic = Anthropic{Mock: false}
	if err := badAnthropic.Validate(); err == nil {
		t.Error("Validate() = nil, want the Anthropic error")
	}
}

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
