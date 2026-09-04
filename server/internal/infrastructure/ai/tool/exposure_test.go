package tool

import (
	"testing"

	"github.com/linzhengen/hub/server/pkg/apicatalog"
)

func defaultCatalog(t *testing.T) *apicatalog.Catalog {
	t.Helper()
	return apicatalog.Default()
}

// TestExcludedRpcsAreNotExposed guards the decision that the permission graph
// and deletions stay out of the assistant's reach.
//
// The two maps are edited by hand and sit next to each other, which is exactly
// the arrangement where an rpc ends up in both. New refuses to build a tool box
// in that case; this fails first, and says which entry is wrong.
func TestExcludedRpcsAreNotExposed(t *testing.T) {
	for full := range escalation {
		if _, listed := exposed[full]; listed {
			t.Errorf("%s is on the excluded list and also on the exposed list", full)
		}
	}
}

// Every exclusion has to name a real rpc, or it guards nothing: a typo or a
// rename leaves an entry that matches no method and an rpc that is quietly
// reachable again.
func TestEveryExclusionNamesARealRpc(t *testing.T) {
	catalog := defaultCatalog(t)
	for full := range escalation {
		if _, ok := catalog.ByFullMethod(full); !ok {
			t.Errorf("%s is excluded but is not an rpc in the API catalog", full)
		}
	}
	for full := range exposed {
		if _, ok := catalog.ByFullMethod(full); !ok {
			t.Errorf("%s is exposed but is not an rpc in the API catalog", full)
		}
	}
}

// TestDecidingIsNeverExposed fixes the asymmetry the request flow rests on.
//
// The assistant may raise an access request and may never decide one. If
// deciding ever became reachable, injected text could travel from a tool result
// to a granted permission inside one conversation, with the same person
// approving it that the text was written to persuade - which is the path the
// escalation list exists to cut.
func TestDecidingIsNeverExposed(t *testing.T) {
	const (
		create = "/system.access.v1.AccessRequestService/CreateAccessRequest"
		decide = "/system.access.v1.AccessRequestService/DecideAccessRequest"
	)

	if !exposed[create] {
		t.Errorf("%s must be exposed as a write: it is the assistant's way into the request flow", create)
	}
	if _, listed := exposed[decide]; listed {
		t.Errorf("%s is exposed; the assistant must never decide a request", decide)
	}
	if !escalation[decide] {
		t.Errorf("%s must be on the escalation list, so leaving it unreachable is a decision and not an omission", decide)
	}
}

// The explain rpcs answer "why can they?" and change nothing, so they are reads.
// A read runs the moment the model asks for it, which is only safe while it
// stays a read.
func TestExplainIsExposedAsARead(t *testing.T) {
	for _, full := range []string{
		"/system.access.v1.AccessService/ExplainUserAccess",
		"/system.access.v1.AccessService/ListPrincipalsForOperation",
		"/system.access.v1.AccessRequestService/ListAccessRequests",
	} {
		write, listed := exposed[full]
		if !listed {
			t.Errorf("%s is not exposed", full)
			continue
		}
		if write {
			t.Errorf("%s is marked as a write; it reads the graph and changes nothing", full)
		}
	}
}

// TestPublicRpcsAreOfferedWithoutAPermission checks that the offer agrees with
// the enforcement.
//
// CreateAccessRequest is public, so the authorization interceptor lets any
// signed-in user call it. Filtering the tool list with Enforce anyway would
// withhold it from precisely the people it exists for - the ones holding no
// permissions at all - and the model would be told it cannot do something the
// server would happily have done.
func TestPublicRpcsAreOfferedWithoutAPermission(t *testing.T) {
	catalog := defaultCatalog(t)
	op, ok := catalog.ByFullMethod("/system.access.v1.AccessRequestService/CreateAccessRequest")
	if !ok {
		t.Fatal("CreateAccessRequest is not in the API catalog")
	}
	if !op.Public {
		t.Error("CreateAccessRequest is no longer public; asking for access now needs a permission, which is a deadlock")
	}
}

// TestMachineIdentityIsNeverExposed fixes the rule that the assistant cannot
// create a principal.
//
// Registering a service account or an agent mints credentials that hold
// permissions with no person attached, which is precisely what the approval
// flow exists to keep out of automation. An agent is the sharper case: it is
// the thing an assistant has the most obvious reason to want, and it will later
// be able to act on a user's behalf, so an assistant able to register one could
// manufacture a second actor and then delegate to it.
//
// Both services are opt-in like everything else, so this test does not change
// behaviour. It records that leaving them out is a decision rather than an
// omission, and fails if a later edit adds one of their rpcs.
func TestMachineIdentityIsNeverExposed(t *testing.T) {
	catalog := defaultCatalog(t)
	for _, service := range []string{
		"ai.agent.v1.AgentService",
		"system.serviceaccount.v1.ServiceAccountService",
	} {
		ops := catalog.OperationsOf(service)
		if len(ops) == 0 {
			t.Errorf("%s is not in the API catalog, so this guard is checking nothing", service)
			continue
		}
		for _, op := range ops {
			if _, listed := exposed[op.FullMethod]; listed {
				t.Errorf("%s is exposed; the assistant must not be able to create or hold a machine identity", op.FullMethod)
			}
		}
	}
}
