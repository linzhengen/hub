package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthRepository is a mock of auth.Repository.
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) Revision(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockAuthRepository) FindUserAuthorizedPolicies(ctx context.Context, userId string) ([]Policy, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Policy), args.Error(1)
}

func (m *MockAuthRepository) FindUserAccessPaths(ctx context.Context, userId string) ([]AccessPath, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]AccessPath), args.Error(1)
}

func (m *MockAuthRepository) FindAccessPaths(ctx context.Context) ([]AccessPath, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]AccessPath), args.Error(1)
}

func (m *MockAuthRepository) FindMemberships(ctx context.Context) ([]Membership, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Membership), args.Error(1)
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
			name:           "Success: wildcard in the middle of the object",
			request:        Request{Subject: subject, Object: "api.system.role.v1.RoleService", Action: "ListRole"},
			mockPolicies:   []Policy{{Object: "api.*.v1.RoleService", Action: "*"}},
			mockError:      nil,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "Success: suffix wildcard on the action",
			request:        Request{Subject: subject, Object: "api.user.v1.UserService", Action: "ListUser"},
			mockPolicies:   []Policy{{Object: "api.user.v1.UserService", Action: "*User"}},
			mockError:      nil,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "Failure: wildcard does not span a missing segment",
			request:        Request{Subject: subject, Object: "api.user.v1.UserService", Action: "ListUser"},
			mockPolicies:   []Policy{{Object: "api.*.v1.RoleService", Action: "*"}},
			mockError:      nil,
			expectedResult: false,
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

// listUser is the operation the explain tests ask about, spelled the way the
// server enforces it.
var listUser = Request{Subject: "user1", Object: "api.user.v1.UserService", Action: "ListUser"}

func TestAuthService_Explain(t *testing.T) {
	ctx := context.Background()

	adminPath := AccessPath{
		GroupId: "group-admin", GroupName: "admin",
		RoleId: "role-admin", RoleName: "admin-role",
		PermissionId: "perm-all", Object: "api.*", Action: "*",
	}
	supportPath := AccessPath{
		GroupId: "group-support", GroupName: "support",
		RoleId: "role-reader", RoleName: "user-reader",
		PermissionId: "perm-list-user", Object: "api.user.v1.UserService", Action: "ListUser",
	}
	unrelatedPath := AccessPath{
		GroupId: "group-support", GroupName: "support",
		RoleId: "role-reader", RoleName: "user-reader",
		PermissionId: "perm-list-role", Object: "api.system.role.v1.RoleService", Action: "ListRole",
	}

	tests := []struct {
		name          string
		mockPaths     []AccessPath
		mockError     error
		expectedPaths []AccessPath
		expectedError bool
	}{
		{
			// The wildcard case is the one a reader most needs told: the answer
			// is yes because of a grant that does not name the operation.
			name:          "a wildcard grant is reported, as the pattern it is",
			mockPaths:     []AccessPath{adminPath},
			expectedPaths: []AccessPath{adminPath},
		},
		{
			// Revoking one of two routes revokes nothing, so both are reported.
			name:          "every route is returned, not the first one found",
			mockPaths:     []AccessPath{adminPath, supportPath},
			expectedPaths: []AccessPath{adminPath, supportPath},
		},
		{
			name:          "routes to other operations are left out",
			mockPaths:     []AccessPath{supportPath, unrelatedPath},
			expectedPaths: []AccessPath{supportPath},
		},
		{
			name:          "a user with no route is explained as having none",
			mockPaths:     []AccessPath{unrelatedPath},
			expectedPaths: nil,
		},
		{
			name:          "repository error",
			mockPaths:     nil,
			mockError:     errors.New("db connection failed"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authRepo := new(MockAuthRepository)
			authRepo.On("FindUserAccessPaths", ctx, listUser.Subject).Return(tt.mockPaths, tt.mockError).Once()

			paths, err := NewService(authRepo).Explain(ctx, listUser)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPaths, paths)
			}
			authRepo.AssertExpectations(t)
		})
	}
}

// TestExplainAgreesWithEnforce is the property that makes an explanation worth
// reading: whenever Enforce says yes, Explain says why, and whenever it says no,
// Explain offers no reason. A second copy of the matching rule would pass its
// own tests and still disagree here.
func TestExplainAgreesWithEnforce(t *testing.T) {
	ctx := context.Background()

	paths := []AccessPath{
		{GroupId: "g1", Object: "api.*", Action: "*"},
		{GroupId: "g2", Object: "api.user.v1.UserService", Action: "*User"},
		{GroupId: "g3", Object: "api.system.*.v1.*Service", Action: "ListRole"},
		{GroupId: "g4", Object: "menu.*", Action: "view"},
	}
	policies := make([]Policy, 0, len(paths))
	for _, path := range paths {
		policies = append(policies, Policy{Object: path.Object, Action: path.Action})
	}

	requests := []Request{
		{Subject: "user1", Object: "api.user.v1.UserService", Action: "ListUser"},
		{Subject: "user1", Object: "api.system.role.v1.RoleService", Action: "ListRole"},
		{Subject: "user1", Object: "api.system.role.v1.RoleService", Action: "DeleteRole"},
		{Subject: "user1", Object: "menu.dashboard", Action: "view"},
		{Subject: "user1", Object: "menu.dashboard", Action: "edit"},
	}

	for _, req := range requests {
		t.Run(req.Object+"/"+req.Action, func(t *testing.T) {
			authRepo := new(MockAuthRepository)
			authRepo.On("FindUserAuthorizedPolicies", ctx, req.Subject).Return(policies, nil).Once()
			authRepo.On("FindUserAccessPaths", ctx, req.Subject).Return(paths, nil).Once()

			service := NewService(authRepo)
			allowed, err := service.Enforce(ctx, req)
			assert.NoError(t, err)
			explained, err := service.Explain(ctx, req)
			assert.NoError(t, err)

			assert.Equal(t, allowed, len(explained) > 0,
				"Enforce answered %v but Explain returned %d routes", allowed, len(explained))
		})
	}
}

func TestAuthService_PrincipalsFor(t *testing.T) {
	ctx := context.Background()
	req := Request{Object: "api.user.v1.UserService", Action: "ListUser"}

	adminPath := AccessPath{
		GroupId: "group-admin", GroupName: "admin",
		RoleId: "role-admin", RoleName: "admin-role",
		PermissionId: "perm-all", Object: "api.*", Action: "*",
	}
	supportPath := AccessPath{
		GroupId: "group-support", GroupName: "support",
		RoleId: "role-reader", RoleName: "user-reader",
		PermissionId: "perm-list-user", Object: "api.user.v1.UserService", Action: "ListUser",
	}
	unrelatedPath := AccessPath{
		GroupId: "group-guest", GroupName: "guest",
		RoleId: "role-guest", RoleName: "guest-role",
		PermissionId: "perm-list-role", Object: "api.system.role.v1.RoleService", Action: "ListRole",
	}

	t.Run("a user holding the operation twice is one principal with both routes", func(t *testing.T) {
		authRepo := new(MockAuthRepository)
		authRepo.On("FindAccessPaths", ctx).
			Return([]AccessPath{adminPath, supportPath, unrelatedPath}, nil).Once()
		authRepo.On("FindMemberships", ctx).Return([]Membership{
			{UserId: "user1", Username: "alice", GroupId: "group-admin"},
			{UserId: "user1", Username: "alice", GroupId: "group-support"},
			{UserId: "user2", Username: "bob", GroupId: "group-support"},
			{UserId: "user3", Username: "carol", GroupId: "group-guest"},
		}, nil).Once()

		principals, err := NewService(authRepo).PrincipalsFor(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, []Principal{
			{UserId: "user1", Username: "alice", Paths: []AccessPath{adminPath, supportPath}},
			{UserId: "user2", Username: "bob", Paths: []AccessPath{supportPath}},
		}, principals)
	})

	t.Run("nobody is allowed, and the members are never read", func(t *testing.T) {
		authRepo := new(MockAuthRepository)
		authRepo.On("FindAccessPaths", ctx).Return([]AccessPath{unrelatedPath}, nil).Once()

		principals, err := NewService(authRepo).PrincipalsFor(ctx, req)

		assert.NoError(t, err)
		assert.Empty(t, principals)
		// Narrowing the graph before the people is the thing that keeps this
		// query from costing users times permissions.
		authRepo.AssertNotCalled(t, "FindMemberships", ctx)
		authRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		authRepo := new(MockAuthRepository)
		authRepo.On("FindAccessPaths", ctx).Return(nil, errors.New("db connection failed")).Once()

		_, err := NewService(authRepo).PrincipalsFor(ctx, req)

		assert.Error(t, err)
		authRepo.AssertExpectations(t)
	})
}

// TestEnforce_ExpiredGrant is the point of the expiry: a grant that has lapsed
// stops working without anybody editing the graph.
func TestEnforce_ExpiredGrant(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	hourAgo := now.Add(-time.Hour)
	hourAway := now.Add(time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		expected  bool
	}{
		{name: "no expiry allows", expiresAt: nil, expected: true},
		{name: "expiry ahead allows", expiresAt: &hourAway, expected: true},
		{name: "expiry passed refuses", expiresAt: &hourAgo, expected: false},
		// The instant itself is out rather than in: a grant "until 17:00" that
		// still worked at 17:00 would be a grant until a moment after it.
		{name: "expiry exactly now refuses", expiresAt: &now, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authRepo := new(MockAuthRepository)
			authRepo.On("FindUserAuthorizedPolicies", ctx, listUser.Subject).Return([]Policy{{
				Object:    listUser.Object,
				Action:    listUser.Action,
				ExpiresAt: tt.expiresAt,
			}}, nil).Once()

			allowed, err := NewServiceWithClock(authRepo, func() time.Time { return now }).
				Enforce(ctx, listUser)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

// TestEnforce_DecidesAgainstTheClockNotTheFetch is the regression test for the
// trap this design exists to avoid.
//
// The policy cache drops what it holds when rbac_revisions moves, and nothing
// moves it when a grant merely runs out - no row is written. So the same
// policies, fetched once, are asked about either side of the expiry. Deciding
// when the policies were fetched would keep the lapsed grant working for the
// length of the cache TTL; deciding against the clock does not.
func TestEnforce_DecidesAgainstTheClockNotTheFetch(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	expiresAt := start.Add(time.Minute)

	// A repository that answers from what it fetched once, as a warm cache does.
	authRepo := new(MockAuthRepository)
	authRepo.On("FindUserAuthorizedPolicies", ctx, listUser.Subject).Return([]Policy{{
		Object:    listUser.Object,
		Action:    listUser.Action,
		ExpiresAt: &expiresAt,
	}}, nil).Twice()

	clock := start
	service := NewServiceWithClock(authRepo, func() time.Time { return clock })

	allowed, err := service.Enforce(ctx, listUser)
	assert.NoError(t, err)
	assert.True(t, allowed, "the grant is live a minute before it expires")

	clock = start.Add(2 * time.Minute)

	allowed, err = service.Enforce(ctx, listUser)
	assert.NoError(t, err)
	assert.False(t, allowed, "the same fetched policies must not outlive the grant")
}

// TestPrincipalsFor_MembershipExpiryEndsTheRoute checks the edge the path query
// never sees: the group may hold the role for a year, but a member who leaves
// on Friday loses the access on Friday.
func TestPrincipalsFor_MembershipExpiryEndsTheRoute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	nextYear := now.AddDate(1, 0, 0)
	friday := now.Add(72 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)
	req := Request{Object: "api.user.v1.UserService", Action: "ListUser"}

	path := AccessPath{
		GroupId: "group-support", GroupName: "support",
		RoleId: "role-reader", RoleName: "user-reader",
		PermissionId: "perm-list-user",
		Object:       req.Object, Action: req.Action,
		ExpiresAt: &nextYear,
	}

	authRepo := new(MockAuthRepository)
	authRepo.On("FindAccessPaths", ctx).Return([]AccessPath{path}, nil).Once()
	authRepo.On("FindMemberships", ctx).Return([]Membership{
		{UserId: "user1", Username: "alice", GroupId: "group-support"},
		{UserId: "user2", Username: "bob", GroupId: "group-support", ExpiresAt: &friday},
		{UserId: "user3", Username: "carol", GroupId: "group-support", ExpiresAt: &yesterday},
	}, nil).Once()

	principals, err := NewServiceWithClock(authRepo, func() time.Time { return now }).
		PrincipalsFor(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, []Principal{
		// alice keeps the group's own expiry.
		{UserId: "user1", Username: "alice", Paths: []AccessPath{path}},
		// bob's membership ends first, so that is when his access ends.
		{UserId: "user2", Username: "bob", Paths: []AccessPath{withExpiryOf(path, &friday)}},
		// carol has already left; she is not a principal at all.
	}, principals)
}

func withExpiryOf(path AccessPath, expiresAt *time.Time) AccessPath {
	path.ExpiresAt = expiresAt
	return path
}

func TestEarliest(t *testing.T) {
	early := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	late := early.Add(time.Hour)

	assert.Nil(t, Earliest(nil, nil), "neither expires, so the route does not")
	assert.Equal(t, &early, Earliest(nil, &early), "nil is never, so it never wins")
	assert.Equal(t, &early, Earliest(&early, nil))
	assert.Equal(t, &early, Earliest(&early, &late))
	assert.Equal(t, &early, Earliest(&late, &early))
}

// TestEnforce_OrganizationScope covers the third term a decision now carries.
//
// The cases are written as a table because they are one rule seen from four
// sides, and the rule is what matters: a grant answers about its own
// organization, the platform's answers about all of them, and a request that
// names none is asking the question every caller asked before organizations
// existed.
func TestEnforce_OrganizationScope(t *testing.T) {
	const (
		orgA = "11111111-1111-1111-1111-111111111111"
		orgB = "22222222-2222-2222-2222-222222222222"
	)

	tests := []struct {
		name    string
		policy  Policy
		reqOrg  string
		allowed bool
	}{
		{
			name:    "a grant held in an organization answers about that organization",
			policy:  Policy{Object: "api.user.v1.UserService", Action: "ListUser", OrgId: orgA},
			reqOrg:  orgA,
			allowed: true,
		},
		{
			name:    "and refuses the same question about another one",
			policy:  Policy{Object: "api.user.v1.UserService", Action: "ListUser", OrgId: orgA},
			reqOrg:  orgB,
			allowed: false,
		},
		{
			name:    "a platform grant answers about any organization",
			policy:  Policy{Object: "api.*", Action: "*", OrgId: "platform", OrgWide: true},
			reqOrg:  orgB,
			allowed: true,
		},
		{
			name:    "a request naming no organization is answered by any grant",
			policy:  Policy{Object: "api.user.v1.UserService", Action: "ListUser", OrgId: orgA},
			reqOrg:  "",
			allowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAuthRepository)
			repo.On("FindUserAuthorizedPolicies", mock.Anything, "user-1").
				Return([]Policy{tt.policy}, nil)

			allowed, err := NewService(repo).Enforce(context.Background(), Request{
				Subject: "user-1",
				Object:  "api.user.v1.UserService",
				Action:  "ListUser",
				OrgId:   tt.reqOrg,
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.allowed, allowed)
		})
	}
}

// TestEnforce_OrganizationAndExpiryComposeIntoOneDecision guards the property
// that makes the two narrowings safe together: a route is live only while every
// edge of it is, so reaching the right organization does not rescue a lapsed
// grant, and being unexpired does not reach the wrong organization.
func TestEnforce_OrganizationAndExpiryComposeIntoOneDecision(t *testing.T) {
	const orgA = "11111111-1111-1111-1111-111111111111"
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	lapsed := now.Add(-time.Hour)

	repo := new(MockAuthRepository)
	repo.On("FindUserAuthorizedPolicies", mock.Anything, "user-1").
		Return([]Policy{{
			Object:    "api.user.v1.UserService",
			Action:    "ListUser",
			OrgId:     orgA,
			ExpiresAt: &lapsed,
		}}, nil)

	svc := NewServiceWithClock(repo, func() time.Time { return now })

	allowed, err := svc.Enforce(context.Background(), Request{
		Subject: "user-1",
		Object:  "api.user.v1.UserService",
		Action:  "ListUser",
		OrgId:   orgA,
	})

	assert.NoError(t, err)
	assert.False(t, allowed, "the right organization does not revive a lapsed grant")
}

// TestExplainNamesTheOrganization checks that an explanation says where, not
// only how. Once two tenants can hold a role of the same name, "they are in
// group X which holds role Y" is not an answer unless X is placed.
func TestExplainNamesTheOrganization(t *testing.T) {
	const orgA = "11111111-1111-1111-1111-111111111111"

	repo := new(MockAuthRepository)
	repo.On("FindUserAccessPaths", mock.Anything, "user-1").
		Return([]AccessPath{{
			GroupId:   "group-1",
			GroupName: "support",
			RoleId:    "role-1",
			RoleName:  "reader",
			Object:    "api.user.v1.UserService",
			Action:    "ListUser",
			OrgId:     orgA,
			OrgName:   "acme",
		}}, nil)

	paths, err := NewService(repo).Explain(context.Background(), Request{
		Subject: "user-1",
		Object:  "api.user.v1.UserService",
		Action:  "ListUser",
		OrgId:   orgA,
	})

	assert.NoError(t, err)
	assert.Len(t, paths, 1)
	assert.Equal(t, orgA, paths[0].OrgId)
	assert.Equal(t, "acme", paths[0].OrgName)
}

// TestVisibleOrgs covers the question a listing asks: not "may they?" but "and
// about whom?".
func TestVisibleOrgs(t *testing.T) {
	const (
		orgA = "11111111-1111-1111-1111-111111111111"
		orgB = "22222222-2222-2222-2222-222222222222"
	)

	platform := Policy{Object: "api.*", Action: "*", OrgId: "platform", OrgWide: true}
	inA := Policy{Object: "api.*", Action: "*", OrgId: orgA}
	inB := Policy{Object: "api.*", Action: "*", OrgId: orgB}

	tests := []struct {
		name     string
		policies []Policy
		asked    string
		want     Scope
	}{
		{
			name:     "a platform grant sees everything",
			policies: []Policy{platform},
			want:     Scope{All: true},
		},
		{
			name:     "a tenant grant sees only that tenant",
			policies: []Policy{inA},
			want:     Scope{OrgIds: []string{orgA}},
		},
		{
			name:     "two tenants are both visible, once each",
			policies: []Policy{inA, inA, inB},
			want:     Scope{OrgIds: []string{orgA, orgB}},
		},
		{
			name:     "naming an organization narrows the platform's view to it",
			policies: []Policy{platform},
			asked:    orgB,
			want:     Scope{OrgIds: []string{orgB}},
		},
		{
			name:     "naming an organization the subject cannot reach admits nothing",
			policies: []Policy{inA},
			asked:    orgB,
			want:     Scope{},
		},
		{
			name:     "holding nothing admits nothing",
			policies: nil,
			want:     Scope{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAuthRepository)
			repo.On("FindUserAuthorizedPolicies", mock.Anything, "user-1").Return(tt.policies, nil)

			scope, err := NewService(repo).VisibleOrgs(context.Background(), "user-1", tt.asked)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.All, scope.All)
			assert.ElementsMatch(t, tt.want.OrgIds, scope.OrgIds)
		})
	}
}

// TestVisibleOrgsDropsALapsedGrant is the property that makes it safe to share a
// cache with Enforce: the scope is decided against the clock, not against
// whenever the policies were fetched.
func TestVisibleOrgsDropsALapsedGrant(t *testing.T) {
	const orgA = "11111111-1111-1111-1111-111111111111"
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	lapsed := now.Add(-time.Hour)

	repo := new(MockAuthRepository)
	repo.On("FindUserAuthorizedPolicies", mock.Anything, "user-1").
		Return([]Policy{{Object: "api.*", Action: "*", OrgId: orgA, ExpiresAt: &lapsed}}, nil)

	svc := NewServiceWithClock(repo, func() time.Time { return now })
	scope, err := svc.VisibleOrgs(context.Background(), "user-1", "")

	assert.NoError(t, err)
	assert.True(t, scope.Empty(), "a lapsed grant shows nothing, as it allows nothing")
}

func TestScopeEmpty(t *testing.T) {
	assert.True(t, Scope{}.Empty())
	assert.False(t, Scope{All: true}.Empty())
	assert.False(t, Scope{OrgIds: []string{"x"}}.Empty())
}
