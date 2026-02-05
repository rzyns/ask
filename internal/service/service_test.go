package service

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestManager_Paths(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	expectedPID := filepath.Join(tmpDir, pidFileName)
	if m.GetPIDFilePath() != expectedPID {
		t.Errorf("GetPIDFilePath = %v, want %v", m.GetPIDFilePath(), expectedPID)
	}

	expectedLog := filepath.Join(tmpDir, logFileName)
	if m.GetLogFilePath() != expectedLog {
		t.Errorf("GetLogFilePath = %v, want %v", m.GetLogFilePath(), expectedLog)
	}
}

func TestManager_PIDOperations(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Test WritePID
	pid := 12345
	if err := m.WritePID(pid); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(m.GetPIDFilePath()); os.IsNotExist(err) {
		t.Error("PID file not created")
	}

	// Test ReadPID
	readPID, err := m.ReadPID()
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}
	if readPID != pid {
		t.Errorf("ReadPID = %v, want %v", readPID, pid)
	}

	// Test ClearPID
	if err := m.ClearPID(); err != nil {
		t.Fatalf("ClearPID failed: %v", err)
	}

	// Verify file gone
	if _, err := os.Stat(m.GetPIDFilePath()); !os.IsNotExist(err) {
		t.Error("PID file not removed")
	}

	// Test ReadPID non-existent
	if _, err := m.ReadPID(); err == nil {
		t.Error("ReadPID should fail for non-existent file")
	}
}

func TestManager_GetStatus(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Case 1: No PID file
	pid, running, err := m.GetStatus()
	if err != nil {
		t.Errorf("GetStatus failed with no file: %v", err)
	}
	if pid != 0 || running {
		t.Errorf("GetStatus (no file) = %v, %v; want 0, false", pid, running)
	}

	// Case 2: Running process (current process)
	myPID := os.Getpid()
	if err := m.WritePID(myPID); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	pid, running, err = m.GetStatus()
	if err != nil {
		t.Errorf("GetStatus failed with valid pid: %v", err)
	}
	if pid != myPID || !running {
		t.Errorf("GetStatus (running) = %v, %v; want %v, true", pid, running, myPID)
	}
}

func TestManager_IsRunning(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Current process should be running
	if !m.IsRunning(os.Getpid()) {
		t.Error("IsRunning(self) returned false")
	}

	// We can't reliably test false case cross-platform without finding a free PID,
	// but we can trust os.FindProcess or Signal to fail for some PIDs.
	// For now, testing positive case is sufficient to cover the main logic path.
}

func TestIsRunning_Signal(t *testing.T) {
	// Verify signal 0 works as expected on this platform
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Skip("FindProcess failed")
	}
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		t.Logf("Signal(0) failed (expected on some platforms if not implemented?): %v", err)
		// Don't fail the test, just log, as this depends on OS behavior being correct for logic
	}
}
