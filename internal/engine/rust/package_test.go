package rust

import (
	"testing"
)

// TestHandleSystemPackage tests deb/rpm/msi handling.
func TestHandleSystemPackage(t *testing.T) {
	e := &RustEngine{}

	// Test deb on non-linux should fail
	err := e.handleSystemPackage(nil, nil, "test", "windows", "x86_64", "", "deb")
	if err == nil {
		t.Error("expected deb on windows to fail")
	}

	// Test msi on linux should fail
	err = e.handleSystemPackage(nil, nil, "test", "linux", "x86_64", "", "msi")
	if err == nil {
		t.Error("expected msi on linux to fail")
	}
}

// TestValidateSystemPackage tests validation logic.
func TestValidateSystemPackage(t *testing.T) {
	e := &RustEngine{}

	// linux + musl should fail
	err := e.validateSystemPackage("linux", "musl", "x86_64", "deb")
	if err == nil {
		t.Error("expected linux+musl to fail")
	}

	// linux + gnu should pass (if cargo-deb is installed, but we can't test that in unit test)
	// Just test the validation logic
}

// TestGetPackageCommand tests command mapping.
func TestGetPackageCommand(t *testing.T) {
	e := &RustEngine{}

	if cmd := e.getPackageCommand("deb"); cmd != "deb" {
		t.Errorf("expected 'deb', got '%s'", cmd)
	}
	if cmd := e.getPackageCommand("rpm"); cmd != "generate-rpm" {
		t.Errorf("expected 'generate-rpm', got '%s'", cmd)
	}
	if cmd := e.getPackageCommand("msi"); cmd != "wix" {
		t.Errorf("expected 'wix', got '%s'", cmd)
	}
	if cmd := e.getPackageCommand("unknown"); cmd != "" {
		t.Errorf("expected '', got '%s'", cmd)
	}
}
