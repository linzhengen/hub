package system

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/server/internal/domain/system/serviceaccount"
	"github.com/linzhengen/hub/server/internal/domain/user"
	"github.com/linzhengen/hub/server/internal/infrastructure/oidc/admin"
)

const (
	creatorID   = "66666666-6666-6666-6666-666666666666"
	keycloakID  = "kc-client-uuid"
	saUserID    = "77777777-7777-7777-7777-777777777777"
	issuedValue = "s3cr3t-value"
)

// fakeOIDCAdmin embeds the real interface so only the three calls a service
// account makes have to be written. Anything else would panic, which is the
// right outcome: this use case has no business touching the rest.
type fakeOIDCAdmin struct {
	admin.Client

	created       admin.ServiceAccountClient
	createErr     error
	createdFor    string
	rotated       string
	rotateErr     error
	deletedClient []string
	deleteErr     error
}

func (f *fakeOIDCAdmin) CreateServiceAccountClient(
	_ context.Context,
	clientID, _ string,
) (admin.ServiceAccountClient, error) {
	f.createdFor = clientID
	if f.createErr != nil {
		return admin.ServiceAccountClient{}, f.createErr
	}
	return f.created, nil
}

func (f *fakeOIDCAdmin) RotateClientSecret(_ context.Context, id string) (string, error) {
	if f.rotateErr != nil {
		return "", f.rotateErr
	}
	f.rotated = id
	return "rotated-value", nil
}

func (f *fakeOIDCAdmin) DeleteClient(_ context.Context, id string) error {
	f.deletedClient = append(f.deletedClient, id)
	return f.deleteErr
}

// fakeAccountRepo records what was stored.
type fakeAccountRepo struct {
	stored    *serviceaccount.ServiceAccount
	createErr error
	findErr   error
	deleted   []string
}

func (f *fakeAccountRepo) Create(_ context.Context, s *serviceaccount.ServiceAccount) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.stored = s
	return nil
}

func (f *fakeAccountRepo) FindOne(_ context.Context, _ string) (*serviceaccount.ServiceAccount, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	copied := *f.stored
	return &copied, nil
}

func (f *fakeAccountRepo) List(
	_ context.Context,
	_ serviceaccount.ListParams,
) ([]*serviceaccount.ServiceAccount, int64, error) {
	return nil, 0, nil
}

func (f *fakeAccountRepo) Delete(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return nil
}

// fakeUserRepo records the `users` row a registration writes, which is the row
// the whole permission model hangs off.
type fakeUserRepo struct {
	user.Repository

	created   *user.User
	createErr error
	deleted   []string
}

func (f *fakeUserRepo) Create(_ context.Context, u *user.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = u
	return nil
}

func (f *fakeUserRepo) Delete(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return nil
}

func newServiceAccountUseCase(
	accounts *fakeAccountRepo,
	users *fakeUserRepo,
	oidc *fakeOIDCAdmin,
) ServiceAccountUseCase {
	return NewServiceAccountUseCase(passthroughTrans{}, accounts, users, oidc)
}

func registeredClient() admin.ServiceAccountClient {
	return admin.ServiceAccountClient{
		KeycloakID: keycloakID,
		UserID:     saUserID,
		Secret:     issuedValue,
	}
}

// TestCreate_WritesTheUserRowKeycloakWillAuthenticateAs is the load-bearing
// test of the whole feature.
//
// A machine authenticates with a client credentials token whose `sub` is the
// Keycloak client's own service account user. The authorization interceptor
// looks the caller up by exactly that id, so unless the `users` row carries it,
// the machine authenticates as nobody: it joins no groups, holds no roles, and
// appears in the audit log as an id nothing matches.
//
// Everything else about service accounts - groups, roles, expiring grants,
// audit - is the ordinary user machinery, and it all hangs off this one field.
func TestCreate_WritesTheUserRowKeycloakWillAuthenticateAs(t *testing.T) {
	accounts := &fakeAccountRepo{}
	users := &fakeUserRepo{}
	oidc := &fakeOIDCAdmin{created: registeredClient()}

	account, credentials, err := newServiceAccountUseCase(accounts, users, oidc).
		Create(ctxAs(creatorID), "ci-deploy", "deploys main")

	require.NoError(t, err)

	require.NotNil(t, users.created)
	assert.Equal(t, saUserID, users.created.Id,
		"the users row must carry the id Keycloak puts in the token's subject")
	// A machine with no mailbox still needs a unique, non-null address, and an
	// inactive row would be refused at the door.
	assert.Equal(t, "service-account-hub-sa-ci-deploy", users.created.Username)
	assert.Equal(t, "hub-sa-ci-deploy@service-account.invalid", users.created.Email)
	assert.Equal(t, user.Active, users.created.Status)

	// The service account row points at the same user, which is what lets the
	// console show a machine and its access as one thing.
	require.NotNil(t, accounts.stored)
	assert.Equal(t, saUserID, accounts.stored.UserId)
	assert.Equal(t, keycloakID, accounts.stored.KeycloakId)
	assert.Equal(t, creatorID, accounts.stored.CreatedByUserId)
	assert.Equal(t, "hub-sa-ci-deploy", accounts.stored.ClientId)
	assert.Equal(t, saUserID, account.UserId)

	// The secret comes back exactly here and nowhere else.
	assert.Equal(t, "hub-sa-ci-deploy", credentials.ClientId)
	assert.Equal(t, issuedValue, credentials.Secret)
}

// The name reaches Keycloak as a client id, so a bad one is refused before the
// round trip rather than after it.
func TestCreate_RefusesABadNameWithoutTouchingKeycloak(t *testing.T) {
	accounts := &fakeAccountRepo{}
	users := &fakeUserRepo{}
	oidc := &fakeOIDCAdmin{created: registeredClient()}

	_, _, err := newServiceAccountUseCase(accounts, users, oidc).
		Create(ctxAs(creatorID), "CI Deploy", "")

	assert.ErrorIs(t, err, errInvalidServiceAccountName)
	assert.Empty(t, oidc.createdFor, "no client may be registered for a name that was refused")
	assert.Nil(t, users.created)
	assert.Nil(t, accounts.stored)
}

func TestCreateServiceAccount_WithoutAUserIsRefused(t *testing.T) {
	oidc := &fakeOIDCAdmin{created: registeredClient()}

	_, _, err := newServiceAccountUseCase(&fakeAccountRepo{}, &fakeUserRepo{}, oidc).
		Create(context.Background(), "ci-deploy", "")

	assert.ErrorIs(t, err, errNoUser)
	assert.Empty(t, oidc.createdFor)
}

// TestCreate_RemovesTheClientWhenItCannotBeRecorded covers the failure that
// would otherwise be the worst of the two: a Keycloak client that works, with
// no row in hub saying it exists. Nobody could find it to turn it off.
func TestCreate_RemovesTheClientWhenItCannotBeRecorded(t *testing.T) {
	for _, tt := range []struct {
		name     string
		accounts *fakeAccountRepo
		users    *fakeUserRepo
	}{
		{
			name:     "the users row fails",
			accounts: &fakeAccountRepo{},
			users:    &fakeUserRepo{createErr: errors.New("insert failed")},
		},
		{
			name:     "the service account row fails",
			accounts: &fakeAccountRepo{createErr: errors.New("insert failed")},
			users:    &fakeUserRepo{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oidc := &fakeOIDCAdmin{created: registeredClient()}

			_, _, err := newServiceAccountUseCase(tt.accounts, tt.users, oidc).
				Create(ctxAs(creatorID), "ci-deploy", "")

			assert.Error(t, err)
			assert.Equal(t, []string{keycloakID}, oidc.deletedClient,
				"a client hub could not record must not be left working")
		})
	}
}

// TestDelete_TakesTheCredentialBeforeTheRecord fixes the order.
//
// Deleting the Keycloak client is what actually stops the machine
// authenticating. Dropping hub's rows first would, on a failure, leave a
// working credential with no record of it - the same state the create path
// goes out of its way to avoid.
func TestDelete_TakesTheCredentialBeforeTheRecord(t *testing.T) {
	accounts := &fakeAccountRepo{stored: serviceaccount.Factory(
		"ci-deploy", "", "hub-sa-ci-deploy", keycloakID, saUserID, creatorID)}
	users := &fakeUserRepo{}
	oidc := &fakeOIDCAdmin{deleteErr: errors.New("keycloak unreachable")}

	err := newServiceAccountUseCase(accounts, users, oidc).Delete(ctxAs(creatorID), "any")

	assert.Error(t, err)
	assert.Empty(t, users.deleted,
		"hub's rows must survive a failure to remove the credential")
}

// Removing the user row is what removes the access: the group memberships and
// the service account row both cascade from it.
func TestDelete_RemovesTheUserTheMachineActedAs(t *testing.T) {
	accounts := &fakeAccountRepo{stored: serviceaccount.Factory(
		"ci-deploy", "", "hub-sa-ci-deploy", keycloakID, saUserID, creatorID)}
	users := &fakeUserRepo{}
	oidc := &fakeOIDCAdmin{}

	err := newServiceAccountUseCase(accounts, users, oidc).Delete(ctxAs(creatorID), "any")

	require.NoError(t, err)
	assert.Equal(t, []string{keycloakID}, oidc.deletedClient)
	assert.Equal(t, []string{saUserID}, users.deleted)
}

func TestRotateSecret_IssuesAgainstTheStoredClient(t *testing.T) {
	accounts := &fakeAccountRepo{stored: serviceaccount.Factory(
		"ci-deploy", "", "hub-sa-ci-deploy", keycloakID, saUserID, creatorID)}
	oidc := &fakeOIDCAdmin{}

	credentials, err := newServiceAccountUseCase(accounts, &fakeUserRepo{}, oidc).
		RotateSecret(ctxAs(creatorID), "any")

	require.NoError(t, err)
	assert.Equal(t, keycloakID, oidc.rotated)
	// The client id is unchanged by a rotation, and is returned alongside so the
	// operator has the whole pair to paste in.
	assert.Equal(t, "hub-sa-ci-deploy", credentials.ClientId)
	assert.Equal(t, "rotated-value", credentials.Secret)
}
