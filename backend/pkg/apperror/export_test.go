package apperror

// Unregister removes a code so a registration test leaves the process-wide
// registry as it found it; without it the docs-sync test would see the test
// code as an undocumented one. It lives in a _test file so production code
// cannot delete a code.
func Unregister(code Code) {
	registryMu.Lock()
	defer registryMu.Unlock()
	delete(registry, code)
}
