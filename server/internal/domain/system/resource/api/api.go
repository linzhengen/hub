package api

// API is one rpc as the RBAC importer sees it.
type API struct {
	// Service is the fully qualified service name, e.g. "user.v1.UserService".
	Service string
	// Method is the rpc name, e.g. "ListUser". It becomes the permission verb
	// unless Action overrides it.
	Method string
	// Resource is the identifier the rpc is guarded by, e.g.
	// "api.user.v1.UserService".
	Resource string
	// Action is the verb the rpc is guarded by. Usually the same as Method.
	Action string
	// Public is true when the rpc skips authorization. A public rpc needs no
	// permission record: nothing would ever read it.
	Public bool
	// Summary describes the rpc, and becomes the permission's description.
	Summary string
}

type APIs []*API
