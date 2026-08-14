package mock_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	chatDomain "github.com/linzhengen/hub/server/internal/domain/ai/chat"
	mockInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/mock"
)

func collect(t *testing.T, ch <-chan chatDomain.Delta) (string, []chatDomain.Delta) {
	t.Helper()
	var text strings.Builder
	var deltas []chatDomain.Delta
	for d := range ch {
		deltas = append(deltas, d)
		text.WriteString(d.Text)
	}
	return text.String(), deltas
}

func TestSendStreamsTheWholeReplyThenDone(t *testing.T) {
	svc := mockInfra.New("claude-mock", 0)

	ch, err := svc.Send(context.Background(), []*chatDomain.Message{
		{Role: chatDomain.RoleUser, Content: "hello there"},
	})
	require.NoError(t, err)

	text, deltas := collect(t, ch)

	require.Greater(t, len(deltas), 1, "the reply must arrive as several deltas so the UI can render a stream")
	require.True(t, deltas[len(deltas)-1].Done, "the last delta closes the stream")
	require.Contains(t, text, "hello there", "the reply echoes the prompt so a developer can tell responses apart")
	require.Contains(t, text, "claude-mock")
	for _, d := range deltas {
		require.NoError(t, d.Error)
	}
}

func TestSendCountsTurns(t *testing.T) {
	svc := mockInfra.New("claude-mock", 0)

	ch, err := svc.Send(context.Background(), []*chatDomain.Message{
		{Role: chatDomain.RoleUser, Content: "first"},
		{Role: chatDomain.RoleAssistant, Content: "answer"},
		{Role: chatDomain.RoleUser, Content: "second"},
	})
	require.NoError(t, err)

	text, _ := collect(t, ch)
	require.Contains(t, text, "turn 2")
	require.Contains(t, text, "second")
	require.NotContains(t, text, "first", "only the latest prompt is echoed")
}

func TestSendReportsAnErrorWhenTheTriggerIsPresent(t *testing.T) {
	svc := mockInfra.New("claude-mock", 0)

	ch, err := svc.Send(context.Background(), []*chatDomain.Message{
		{Role: chatDomain.RoleUser, Content: "please !error now"},
	})
	require.NoError(t, err)

	_, deltas := collect(t, ch)

	require.Len(t, deltas, 1)
	require.Error(t, deltas[0].Error)
	require.False(t, deltas[0].Done, "an error ends the stream instead of completing it")
}

func TestSendStopsWhenTheContextIsCancelled(t *testing.T) {
	svc := mockInfra.New("claude-mock", 0)

	ctx, cancel := context.WithCancel(context.Background())
	ch, err := svc.Send(ctx, []*chatDomain.Message{
		{Role: chatDomain.RoleUser, Content: "hello"},
	})
	require.NoError(t, err)
	cancel()

	// The channel must still be closed, otherwise the use case's reader leaks.
	for range ch { //nolint:revive // draining is the point
	}
}

func TestSendWithoutAUserMessage(t *testing.T) {
	svc := mockInfra.New("claude-mock", 0)

	ch, err := svc.Send(context.Background(), nil)
	require.NoError(t, err)

	text, deltas := collect(t, ch)
	require.NotEmpty(t, text)
	require.True(t, deltas[len(deltas)-1].Done)
}
