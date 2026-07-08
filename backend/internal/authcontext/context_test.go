package authcontext

import (
	"context"
	"testing"
	"time"
)

func TestUserID(t *testing.T) {
	ctx := WithUserID(testContext{}, 123)
	id, ok := UserID(ctx)
	if !ok || id != 123 {
		t.Fatalf("expected user id 123, got %v %v", id, ok)
	}
}

type testContext struct{}

func (testContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (testContext) Done() <-chan struct{}       { return nil }
func (testContext) Err() error                  { return nil }
func (testContext) Value(key any) any           { return nil }

var _ context.Context = testContext{}
