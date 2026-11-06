package queue

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"

	"server/internal/tasks"
)

var _ Queue = (*asynqQueue)(nil)

func TestQueue_SendEmail_EnqueuesEmailTask(t *testing.T) {
	mr := miniredis.RunT(t)

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: mr.Addr()})
	defer client.Close()

	q := New(client)

	payload := tasks.EmailPayload{
		FromEmail: "hello@example.com",
		ToEmail:   []string{"user1@example.com", "user2@example.com"},
		Subject:   "Welcome to Trimium!",
		HTMLBody:  "<h1>Welcome!</h1>",
	}

	if err := q.SendEmail(t.Context(), payload); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr()})
	defer inspector.Close()

	infos, err := inspector.ListPendingTasks("default")
	if err != nil {
		t.Fatalf("failed to list pending tasks: %v", err)
	}

	if len(infos) != 1 {
		t.Fatalf("expected 1 pending task, got %d", len(infos))
	}

	if infos[0].Type != tasks.TypeEmailSend {
		t.Errorf("expected task type %q, got %q", tasks.TypeEmailSend, infos[0].Type)
	}

	var got tasks.EmailPayload
	if err := json.Unmarshal(infos[0].Payload, &got); err != nil {
		t.Fatalf("failed to decode enqueued payload: %v", err)
	}

	if !reflect.DeepEqual(got, payload) {
		t.Errorf("enqueued payload mismatch: got %#v, want %#v", got, payload)
	}
}

func TestQueue_SendEmail_FailsOnClosedClient(t *testing.T) {
	mr := miniredis.RunT(t)

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: mr.Addr()})
	if err := client.Close(); err != nil {
		t.Fatalf("failed to close client: %v", err)
	}

	err := New(client).SendEmail(t.Context(), tasks.EmailPayload{})
	if err == nil {
		t.Fatal("expected an error when enqueueing on a closed client")
	}
	if errors.Is(err, context.Canceled) {
		t.Errorf("expected an enqueue error, got %v", err)
	}
}
