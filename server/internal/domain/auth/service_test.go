package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthRepository is a mock of auth.Repository.
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) FindUserAuthorizedPolicies(ctx context.Context, userId string) ([]Policy, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Policy), args.Error(1)
}

func TestAuthService_Enforce(t *testing.T) {
	ctx := context.Background()
	subject := "user1"

	tests := []struct {
		name           string
		request        Request
		mockPolicies   []Policy
		mockError      error
		expectedResult bool
		expectedError  bool
	}{
		{
			name:           "Success: exact match",
			request:        Request{Subject: subject, Object: "articles", Action: "read"},
			mockPolicies:   []Policy{{Object: "articles", Action: "read"}},
			mockError:      nil,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "Success: prefix wildcard match",
			request:        Request{Subject: subject, Object: "articles:123", Action: "write"},
			mockPolicies:   []Policy{{Object: "articles:*", Action: "write"}},
			mockError:      nil,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "Success: full wildcard match",
			request:        Request{Subject: subject, Object: "any_object", Action: "any_action"},
			mockPolicies:   []Policy{{Object: "*", Action: "*"}},
			mockError:      nil,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "Failure: action mismatch",
			request:        Request{Subject: subject, Object: "articles", Action: "delete"},
			mockPolicies:   []Policy{{Object: "articles", Action: "write"}},
			mockError:      nil,
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:           "Failure: object mismatch",
			request:        Request{Subject: subject, Object: "users", Action: "read"},
			mockPolicies:   []Policy{{Object: "articles", Action: "read"}},
			mockError:      nil,
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:           "Expect false: no policies match",
			request:        Request{Subject: subject, Object: "articles", Action: "read"},
			mockPolicies:   []Policy{},
			mockError:      nil,
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:           "Failure: repository returns error",
			request:        Request{Subject: subject, Object: "articles", Action: "read"},
			mockPolicies:   nil,
			mockError:      errors.New("db connection failed"),
			expectedResult: false,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authRepo := new(MockAuthRepository)
			authRepo.On("FindUserAuthorizedPolicies", ctx, tt.request.Subject).Return(tt.mockPolicies, tt.mockError).Once()

			service := NewService(authRepo)
			result, err := service.Enforce(ctx, tt.request)

			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			authRepo.AssertExpectations(t)
		})
	}
}
