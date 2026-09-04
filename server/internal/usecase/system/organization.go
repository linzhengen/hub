package system

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/system/organization"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/postgres"
	"github.com/linzhengen/hub/server/internal/usecase/pagination"
	"github.com/linzhengen/hub/server/internal/usecase/scope"
	"github.com/linzhengen/hub/server/pkg/logger"
)

var (
	errInvalidOrganizationSlug = status.Error(
		codes.InvalidArgument,
		"an organization slug must be 3-64 characters of lowercase letters, digits and hyphens, "+
			"starting and ending with a letter or digit",
	)
	// errPlatformOrganization refuses the two edits that would rewrite every
	// decision in the installation at once.
	errPlatformOrganization = status.Error(
		codes.FailedPrecondition,
		"the platform organization cannot be deleted: every group that predates organizations belongs to it",
	)
	errNoCaller = status.Error(codes.Unauthenticated, "user not authenticated")
)

// OrganizationUseCase manages the tenants of this installation.
//
// An organization is a boundary rather than a container: nothing is stored
// "inside" one, groups simply say which they belong to. That is why there is no
// membership rpc here - a user joins an organization by joining one of its
// groups, through the same AddUsersToGroup that has always existed.
type OrganizationUseCase interface {
	Create(ctx context.Context, o *organization.Organization) (*organization.Organization, error)
	Get(ctx context.Context, id string) (*organization.Organization, error)
	List(ctx context.Context, params organization.ListParams) ([]*organization.Organization, int64, error)
	Update(ctx context.Context, o *organization.Organization) (*organization.Organization, error)
	Delete(ctx context.Context, id string) error
	// ListMine is the organizations the calling user belongs to.
	ListMine(ctx context.Context) ([]*organization.Organization, error)
}

func NewOrganizationUseCase(
	db *sql.DB,
	dialectWrapper persistence.DialectWrapper,
	transRepo trans.Repository,
	orgRepo organization.Repository,
	authSvc auth.Service,
) OrganizationUseCase {
	return &organizationUseCase{
		db:             db,
		dialectWrapper: dialectWrapper,
		transRepo:      transRepo,
		orgRepo:        orgRepo,
		authSvc:        authSvc,
	}
}

type organizationUseCase struct {
	db             *sql.DB
	dialectWrapper persistence.DialectWrapper
	transRepo      trans.Repository
	orgRepo        organization.Repository
	authSvc        auth.Service
}

func (uc organizationUseCase) Create(
	ctx context.Context,
	o *organization.Organization,
) (*organization.Organization, error) {
	if !organization.ValidSlug(o.Slug) {
		return nil, errInvalidOrganizationSlug
	}
	// A second PLATFORM organization would hold every tenant's data, so the kind
	// is refused in the proto and again here: validation states the rule for
	// callers, this states it for every other path into the use case.
	if o.Kind.AppliesEverywhere() {
		return nil, status.Error(codes.InvalidArgument, "the platform organization already exists and there is only one")
	}
	if err := uc.orgRepo.Create(ctx, o); err != nil {
		return nil, err
	}
	return uc.orgRepo.FindOne(ctx, o.Id)
}

func (uc organizationUseCase) Get(ctx context.Context, id string) (*organization.Organization, error) {
	return uc.orgRepo.FindOne(ctx, id)
}

// List reads the organizations the caller can reach.
//
// The query lives here rather than in the repository for the same reason the
// group and role listings do: it is a read model with filters, and one of those
// filters is a set whose size is not known until the caller's scope is.
func (uc organizationUseCase) List(
	ctx context.Context,
	params organization.ListParams,
) ([]*organization.Organization, int64, error) {
	// Listing every tenant would tell one customer who the others are, which is
	// the disclosure a boundary is for even when no other data crosses it.
	visible, err := scope.VisibleOrgs(ctx, uc.authSvc)
	if err != nil {
		return nil, 0, err
	}
	if visible.Empty() {
		return nil, 0, nil
	}

	b := uc.dialectWrapper.From("organizations")
	if !visible.All {
		b = b.Where(goqu.Ex{"id": visible.OrgIds})
	}
	if params.Ids != nil {
		b = b.Where(postgres.In("id", params.Ids))
	}
	if params.Name != "" {
		b = b.Where(goqu.C("name").ILike(fmt.Sprintf("%%%s%%", params.Name)))
	}
	if params.Slug != "" {
		b = b.Where(goqu.C("slug").ILike(fmt.Sprintf("%%%s%%", params.Slug)))
	}
	if params.Kind != "" {
		b = b.Where(goqu.Ex{"kind": string(params.Kind)})
	}

	cnt, err := postgres.SelectCount(ctx, uc.db, b)
	if err != nil {
		return nil, 0, err
	}

	page := pagination.New(params.Limit, params.Offset)
	b = b.Order(goqu.C("name").Asc()).Limit(page.Limit()).Offset(page.Offset())

	items, err := uc.list(ctx, b)
	if err != nil {
		return nil, 0, err
	}
	return items, cnt, nil
}

func (uc organizationUseCase) list(
	ctx context.Context,
	b *goqu.SelectDataset,
) ([]*organization.Organization, error) {
	query, queryParams, err := b.Select("*").Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := uc.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Infof("error closing rows: %v", err)
		}
	}()

	items := make([]*organization.Organization, 0)
	for rows.Next() {
		var (
			o    organization.Organization
			kind string
			st   string
		)
		// Scanned positionally from SELECT *, so the order follows 000002.
		if err := rows.Scan(
			&o.Id,
			&o.Name,
			&o.Slug,
			&kind,
			&o.Description,
			&st,
			&o.CreatedAt,
			&o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		o.Kind = organization.Kind(kind)
		o.Status = organization.Status(st)
		items = append(items, &o)
	}
	return items, rows.Err()
}

// Update changes what an organization is called and whether it is active. The
// kind is not among them: see the note on UpdateOrganization in organization.sql.
func (uc organizationUseCase) Update(
	ctx context.Context,
	o *organization.Organization,
) (*organization.Organization, error) {
	if !organization.ValidSlug(o.Slug) {
		return nil, errInvalidOrganizationSlug
	}
	if err := uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		existing, err := uc.orgRepo.FindOne(ctx, o.Id)
		if err != nil {
			return err
		}
		// The kind is carried over from the stored row rather than taken from
		// the caller, so an update can never change it even by accident.
		o.Kind = existing.Kind
		return uc.orgRepo.Update(ctx, o)
	}); err != nil {
		return nil, err
	}
	return uc.orgRepo.FindOne(ctx, o.Id)
}

// Delete removes an organization and, by cascade, every group in it.
//
// The platform organization is refused. Deleting it would take every group the
// seed creates with it - `admin` among them - and revoke every administrator in
// one statement.
func (uc organizationUseCase) Delete(ctx context.Context, id string) error {
	if id == organization.PlatformOrgId {
		return errPlatformOrganization
	}
	return uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		existing, err := uc.orgRepo.FindOne(ctx, id)
		if err != nil {
			return err
		}
		if existing.Platform() {
			return errPlatformOrganization
		}
		return uc.orgRepo.Delete(ctx, id)
	})
}

func (uc organizationUseCase) ListMine(ctx context.Context) ([]*organization.Organization, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoCaller
	}
	return uc.orgRepo.FindByUser(ctx, userId)
}
