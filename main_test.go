package main

import (
	"testing"
	"time"
)

func TestHandleSigTerms(t *testing.T) {
	// This test is tricky because handleSigTerms starts a goroutine
	// and we can't easily test os.Exit() calls
	// We'll test that the function doesn't panic when called

	// Call handleSigTerms - it should start a goroutine and return immediately
	handleSigTerms()

	// Give the goroutine a moment to start
	time.Sleep(10 * time.Millisecond)

	// The function should have returned without error
	// We can't easily test the signal handling without complex setup
	// but we can verify the function is callable
}

func TestMain(t *testing.T) {
	// We can't easily test the main function directly because it calls os.Exit
	// and cmd.Execute() which would require complex mocking
	// Instead, we'll test that we can call handleSigTerms without panic
	handleSigTerms()
}
