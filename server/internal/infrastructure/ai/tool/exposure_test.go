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
