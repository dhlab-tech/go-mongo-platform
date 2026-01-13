# API Stability Policy

**Status:** Public Normative Document  
**Version:** 1.0  
**Purpose:** Define public API boundaries and stability guarantees

---

## Public API vs Internal

### Public API

The following packages are considered **public API**:

- `pkg/mongo/**` — MongoDB operations and Change Streams integration
- `pkg/inmemory/**` — In-memory projection and indexing (exported types and functions)

**Public API guarantees:**
- Breaking changes are documented
- Deprecation warnings are provided before removal
- Semantic versioning applies

### Internal API

Everything else is considered **internal**:

- `internal/**` (if exists)
- Unexported types and functions
- Test files
- Examples

**Internal API:**
- No stability guarantees
- May change without notice
- Not intended for external use

---

## Versioning Policy

### Pre-1.0 (v0.x)

**Current status:** v0.1.0 (pre-1.0)

**Policy:**
- Breaking changes are **allowed** but must be documented
- Breaking changes should be justified and communicated
- Deprecation warnings should be provided when possible
- Minor version increments (v0.1.0 → v0.2.0) may include breaking changes
- Patch version increments (v0.1.0 → v0.1.1) should not include breaking changes

**Semver:** Pre-1.0 semver (breaking changes allowed)

---

### Post-1.0 (v1.x+)

**Policy (after v1.0 release):**
- **Strict semantic versioning:**
  - **MAJOR** (v1.0.0 → v2.0.0): Breaking changes
  - **MINOR** (v1.0.0 → v1.1.0): New features, backward compatible
  - **PATCH** (v1.0.0 → v1.0.1): Bug fixes, backward compatible
- Breaking changes require major version increment
- Deprecation warnings must be provided at least one minor version before removal

**Semver:** Strict semantic versioning

---

## Deprecation Policy

### Deprecation Process

1. **Announcement:** Deprecated APIs are marked with Go doc comment:
   ```go
   // Deprecated: Use NewFunction instead.
   func OldFunction() { ... }
   ```

2. **Timeline:** 
   - Pre-1.0: Deprecation warnings provided when possible
   - Post-1.0: Deprecation warnings provided at least one minor version before removal

3. **Removal:**
   - Pre-1.0: May be removed in next minor version
   - Post-1.0: Removed only in next major version

### Deprecation Examples

**Good:**
```go
// Deprecated: Use NewProcessor instead. OldProcessor will be removed in v0.2.0.
func OldProcessor() { ... }
```

**Better (post-1.0):**
```go
// Deprecated: Use NewProcessor instead. OldProcessor will be removed in v2.0.0.
func OldProcessor() { ... }
```

---

## Breaking Changes

### What Constitutes a Breaking Change

- Removing exported functions or types
- Changing function signatures (parameters, return types)
- Changing behavior in a way that breaks existing code
- Removing package exports
- Changing struct field types or removing fields

### What Does NOT Constitute a Breaking Change

- Adding new functions or types
- Adding optional parameters (with defaults)
- Bug fixes that correct incorrect behavior
- Performance improvements
- Internal implementation changes

---

## Migration Guide

When breaking changes occur:

1. **Deprecation notice** (when possible)
2. **Migration guide** in CHANGELOG.md
3. **Examples** showing old vs. new usage
4. **Clear communication** in release notes

---

## Examples and Documentation

- **Examples:** May change without notice (not part of public API)
- **Documentation:** May be updated to reflect best practices
- **README.md:** Public-facing, but structure may evolve

---

## Reporting API Issues

If you encounter:
- Unexpected breaking changes
- Missing deprecation warnings
- Unclear API boundaries

Please report via:
- GitHub Issues
- GitHub Discussions

---

## Related Documentation

- [CHANGELOG.md](CHANGELOG.md) — Version history and breaking changes
- [ARCHITECTURE_CONTRACT.md](ARCHITECTURE_CONTRACT.md) — Architectural constraints
- [CONTRIBUTING.md](CONTRIBUTING.md) — Contributing guidelines

---

**End of API Stability Policy**

