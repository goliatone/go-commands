package cron

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/goliatone/command"
)

// Test message type
type TestMessage struct{}

func (m TestMessage) Type() string { return "test.message" }

// Test command handler
type TestCommandHandler struct {
	executionCount int
	mu             sync.Mutex
}

func (h *TestCommandHandler) Execute(ctx context.Context, msg command.Message) error {
	h.mu.Lock()
	h.executionCount++
	h.mu.Unlock()
	return nil
}

func (h *TestCommandHandler) GetExecutionCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.executionCount
}

// Test query handler
type TestQueryHandler struct {
	queryCount int
	mu         sync.Mutex
}

func (h *TestQueryHandler) Query(ctx context.Context, msg command.Message) (any, error) {
	h.mu.Lock()
	h.queryCount++
	count := h.queryCount
	h.mu.Unlock()
	return count, nil
}

func (h *TestQueryHandler) GetQueryCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.queryCount
}

func TestCronScheduler(t *testing.T) {
	t.Run("command handler execution", func(t *testing.T) {
		scheduler := NewCronScheduler()
		handler := &TestCommandHandler{}

		// Schedule job to run every second
		entryID, err := scheduler.AddHandler(HandlerOptions{
			Expression: "@every 1s",
		}, handler)

		if err != nil {
			t.Fatalf("Failed to add handler: %v", err)
		}

		scheduler.Start()
		time.Sleep(1500 * time.Millisecond)
		scheduler.Stop()

		if count := handler.GetExecutionCount(); count == 0 {
			t.Error("Command handler was not executed")
		}

		scheduler.RemoveHandler(entryID)
	})

	t.Run("command handler execution", func(t *testing.T) {
		scheduler := NewCronScheduler(WithParser(SecondsParser))
		handler := &TestCommandHandler{}

		// Schedule job to run every second
		entryID, err := scheduler.AddHandler(HandlerOptions{
			Expression: "* * * * * *",
		}, handler)

		if err != nil {
			t.Fatalf("Failed to add handler: %v", err)
		}

		scheduler.Start()
		time.Sleep(1500 * time.Millisecond)
		scheduler.Stop()

		if count := handler.GetExecutionCount(); count == 0 {
			t.Error("Command handler was not executed")
		}

		scheduler.RemoveHandler(entryID)
	})

	t.Run("function handler execution", func(t *testing.T) {
		count := 0
		scheduler := NewCronScheduler()
		handler := func() {
			count = count + 1
		}

		// Schedule job to run every second
		entryID, err := scheduler.AddHandler(HandlerOptions{
			Expression: "@every 1s",
		}, handler)

		if err != nil {
			t.Fatalf("Failed to add handler: %v", err)
		}

		scheduler.Start()
		// Wait for at least one execution
		time.Sleep(2 * time.Second)
		scheduler.Stop()

		if count == 0 {
			t.Error("Query handler was not executed")
		}

		scheduler.RemoveHandler(entryID)
	})

	t.Run("invalid cron expression", func(t *testing.T) {
		scheduler := NewCronScheduler()
		handler := &TestCommandHandler{}

		_, err := scheduler.AddHandler(HandlerOptions{
			Expression: "invalid",
		}, handler)

		if err == nil {
			t.Error("Expected error for invalid cron expression")
		}
	})

	t.Run("empty cron expression", func(t *testing.T) {
		scheduler := NewCronScheduler()
		handler := &TestCommandHandler{}

		_, err := scheduler.AddHandler(HandlerOptions{
			Expression: "",
		}, handler)

		if err == nil {
			t.Error("Expected error for empty cron expression")
		}
	})

	t.Run("invalid handler type", func(t *testing.T) {
		scheduler := NewCronScheduler()
		handler := struct{}{}

		_, err := scheduler.AddHandler(HandlerOptions{
			Expression: "* * * * *",
		}, handler)

		if err == nil {
			t.Error("Expected error for invalid handler type")
		}
	})
}