package app

import (
	"context"
	"testing"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Error("NewApp() returned nil")
	}
}

func TestApp_Startup(t *testing.T) {
	app := NewApp()
	ctx := context.TODO()
	app.Startup(ctx)
	if app.ctx != ctx {
		t.Error("Startup did not save the context")
	}
}
