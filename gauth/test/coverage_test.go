package test

import (
	"testing"

	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
)

// TestMain sets up test environment and runs coverage
func TestMain(m *testing.M) {
	// This file ensures that test coverage includes the test package
	// The actual TestMain is in individual test packages
	m.Run()
}

// TestCoverageCompilation ensures all test files compile correctly
func TestCoverageCompilation(t *testing.T) {
	t.Log("Test coverage compilation check")

	// Basic test to ensure test helpers work
	testDB := testhelpers.SetupTestDB(t)
	defer testDB.Cleanup()

	if testDB.GetDB() == nil {
		t.Fatal("Test database setup failed")
	}

	t.Log("Test helpers working correctly")
}
