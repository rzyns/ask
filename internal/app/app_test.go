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

func TestApp_Greet(t *testing.T) {
	app := NewApp()
	name := "Tester"
	expected := "Hello Tester, It's show time!"
	result := app.Greet(name)
	if result != expected {
		t.Errorf("Greet(%s) = %s; want %s", name, result, expected)
	}
}
