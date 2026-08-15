package system

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/system/serviceaccount"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/domain/user"
	"github.com/linzhengen/hub/server/internal/infrastructure/oidc/admin"
	"github.com/linzhengen/hub/server/internal/usecase/pagination"
	"github.com/linzhengen/hub/server/pkg/logger"
)

var errInvalidServiceAccountName = status.Error(
	codes.InvalidArgument,
	"a service account name must be 3-64 characters of lowercase letters, digits and hyphens, "+
		"starting and ending with a letter or digit",
)

// Credentials are what a machine needs to authenticate, returned once.
//
// The secret is not stored by hub and cannot be read back: Keycloak can issue a
// new one but not show the current one again in every configuration, so
// write-once is the behaviour that holds everywhere. Losing it means rotating
// it, which is the right cost for a credential.
type Credentials struct {
	ClientId string
	Secret   string
}

// ServiceAccountUseCase registers the machines that call the API.
//
// It does not authenticate them. A service account is a Keycloak client and a
// machine gets its token through the client credentials grant exactly as before;
// what this adds is that hub knows the client exists, what it is for, who
// created it, and which hub user it acts as.
type ServiceAccountUseCase interface {
	Create(ctx context.Context, name, description string) (*serviceaccount.ServiceAccount, Credentials, error)
	Get(ctx context.Context, id string) (*serviceaccount.ServiceAccount, error)
	List(ctx context.Context, params serviceaccount.ListParams) ([]*serviceaccount.ServiceAccount, int64, error)
	RotateSecret(ctx context.Context, id string) (Credentials, error)
	Delete(ctx context.Context, id string) error
}

func NewServiceAccountUseCase(
	transRepo trans.Repository,
	accountRepo serviceaccount.Repository,
	userRepo user.Repository,
	oidcAdmin admin.Client,
) ServiceAccountUseCase {
	return &serviceAccountUseCase{
		transRepo:   transRepo,
		accountRepo: accountRepo,
		userRepo:    userRepo,
		oidcAdmin:   oidcAdmin,
	}
}

type serviceAccountUseCase struct {
	transRepo   trans.Repository
	accountRepo serviceaccount.Repository
	userRepo    user.Repository
	oidcAdmin   admin.Client
}

// Create registers a machine and returns its credentials once.
//
// The Keycloak client is made first, because hub's rows are derived from what
// it returns - the id of the user the client acts as, above all. If storing
// those rows then fails, the client is removed again: a client nobody has a
// record of is a working credential nobody can see.
func (uc serviceAccountUseCase) Create(
	ctx context.Context,
	name, description string,
) (*serviceaccount.ServiceAccount, Credentials, error) {
	creator, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, Credentials{}, errNoUser
	}
	if !serviceaccount.ValidName(name) {
		return nil, Credentials{}, errInvalidServiceAccountName
	}

	clientId := serviceaccount.ClientIdFor(name)
	created, err := uc.oidcAdmin.CreateServiceAccountClient(ctx, clientId, description)
	if err != nil {
		return nil, Credentials{}, err
	}

	account := serviceaccount.Factory(
		name, description, clientId, created.KeycloakID, created.UserID, creator)

	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		// The machine gets a row in `users` like anybody else, so that groups,
		// roles and the audit log all work on it unchanged. It is created here
		// rather than left to the interceptor's provisioning: a client
		// credentials token carries no email, and `users.email` is unique and
		// not null, so the first machine would take the empty address and the
		// second would collide with it.
		if err := uc.userRepo.Create(ctx, &user.User{
			Id:        created.UserID,
			Username:  serviceaccount.UsernameFor(name),
			Email:     serviceaccount.EmailFor(name),
			Status:    user.Active,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}); err != nil {
			return err
		}
		return uc.accountRepo.Create(ctx, account)
	}); err != nil {
		uc.removeClient(ctx, created.KeycloakID)
		return nil, Credentials{}, err
	}

	return account, Credentials{ClientId: clientId, Secret: created.Secret}, nil
}

func (uc serviceAccountUseCase) Get(
	ctx context.Context,
	id string,
) (*serviceaccount.ServiceAccount, error) {
	return uc.accountRepo.FindOne(ctx, id)
}

func (uc serviceAccountUseCase) List(
	ctx context.Context,
	params serviceaccount.ListParams,
) ([]*serviceaccount.ServiceAccount, int64, error) {
	page := pagination.New(params.Limit, params.Offset)
	params.Limit = uint32(page.Limit())   //nolint:gosec // pagination bounds it
	params.Offset = uint32(page.Offset()) //nolint:gosec // pagination bounds it
	return uc.accountRepo.List(ctx, params)
}

// RotateSecret issues a new secret and invalidates the old one, which is the
// only remedy for one that leaked or was lost.
func (uc serviceAccountUseCase) RotateSecret(ctx context.Context, id string) (Credentials, error) {
	account, err := uc.accountRepo.FindOne(ctx, id)
	if err != nil {
		return Credentials{}, err
	}

	secret, err := uc.oidcAdmin.RotateClientSecret(ctx, account.KeycloakId)
	if err != nil {
		return Credentials{}, err
	}
	return Credentials{ClientId: account.ClientId, Secret: secret}, nil
}

// Delete removes the machine's identity entirely.
//
// The Keycloak client goes first, because that is what actually stops the
// machine authenticating: dropping hub's rows while leaving the client would
// take the record away and leave the credential working, which is the worse of
// the two failures to be left with.
//
// Removing the user row cascades to its group memberships and to the service
// account row, so nothing is left holding permissions.
func (uc serviceAccountUseCase) Delete(ctx context.Context, id string) error {
	account, err := uc.accountRepo.FindOne(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.oidcAdmin.DeleteClient(ctx, account.KeycloakId); err != nil {
		return err
	}

	return uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		return uc.userRepo.Delete(ctx, account.UserId)
	})
}

// removeClient undoes a client that hub could not record. It is best effort:
// the caller is already returning an error, and there is nothing useful to do
// with a second one except say so.
func (uc serviceAccountUseCase) removeClient(ctx context.Context, keycloakId string) {
	if err := uc.oidcAdmin.DeleteClient(ctx, keycloakId); err != nil {
		logger.Errorf(
			"service account: could not remove keycloak client %s after failing to record it; "+
				"it exists in keycloak with no row in hub: %v",
			keycloakId, err)
	}
}
