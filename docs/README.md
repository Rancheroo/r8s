# r8s Documentation Index

This directory contains all project documentation organized by category.

## Core Documentation (Root Level)

These essential files remain in the project root for easy discovery:

- **[README.md](../README.md)** - Project overview and quick start
- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - Development guide
- **[LESSONS-LEARNED.md](../LESSONS-LEARNED.md)** - Project wisdom and principles
- **[TROUBLESHOOTING.md](../TROUBLESHOOTING.md)** - Common issues and solutions

## Documentation Structure

### `/docs/` - User Documentation
- **[USAGE.md](USAGE.md)** - Complete CLI reference
- **[BUNDLE-FORMAT.md](BUNDLE-FORMAT.md)** - RKE2 bundle structure
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Technical design
- **[CHANGELOG.md](CHANGELOG.md)** - Version history

### `/docs/development/` - Development Resources
- **[PRINCIPLES.md](development/PRINCIPLES.md)** - Design principles
- **[FUTURE_WORK.md](development/FUTURE_WORK.md)** - Roadmap and future features
- **[POST_MERGE_CLEANUP.md](development/POST_MERGE_CLEANUP.md)** - Branch cleanup tasks
- **[V0.6.8_PROMPT.md](development/V0.6.8_PROMPT.md)** - v0.6.8 development prompt

### `/docs/test-reports/` - Test Documentation
- **[V0.5.5_FEATURE1_TEST_PLAN.md](test-reports/V0.5.5_FEATURE1_TEST_PLAN.md)**
- **[V0.5.5_KNOWN_LIMITATIONS.md](test-reports/V0.5.5_KNOWN_LIMITATIONS.md)**
- **[V0.6.2_TEST_SUMMARY.md](test-reports/V0.6.2_TEST_SUMMARY.md)**

### `/docs/hotfix-reports/` - Hotfix Documentation
- **[HOTFIX_V0.6.8.1_FINAL_SUCCESS.md](hotfix-reports/HOTFIX_V0.6.8.1_FINAL_SUCCESS.md)** - Final success report
- **[HOTFIX_V0.6.8.1_RETEST_FAILED.md](hotfix-reports/HOTFIX_V0.6.8.1_RETEST_FAILED.md)** - Round 2 test results
- **[HOTFIX_V0.6.8.1_TEST_RESULTS.md](hotfix-reports/HOTFIX_V0.6.8.1_TEST_RESULTS.md)** - Round 1 test results

### `/docs/archive/` - Historical Documentation
Older development notes, test reports, and completed work organized by date and phase.

## Quick Links

### For Users
- [Quick Start](../README.md#quick-start)
- [Features](../README.md#features)
- [Keyboard Shortcuts](../README.md#keyboard-shortcuts)
- [Troubleshooting](../TROUBLESHOOTING.md)
- [CLI Reference](USAGE.md)

### For Contributors
- [Contributing Guide](../CONTRIBUTING.md)
- [Architecture](ARCHITECTURE.md)
- [Design Principles](development/PRINCIPLES.md)
- [Lessons Learned](../LESSONS-LEARNED.md)
- [Development Roadmap](development/FUTURE_WORK.md)

### For Maintainers
- [Changelog](CHANGELOG.md)
- [Hotfix Reports](hotfix-reports/)
- [Test Reports](test-reports/)
- [Post-Merge Tasks](development/POST_MERGE_CLEANUP.md)

## Documentation Principles

1. **Keep root clean** - Only essential user-facing docs in project root
2. **Organize by purpose** - Group related docs in subdirectories
3. **Archive when done** - Move completed work to `/docs/archive/`
4. **Date archives** - Use YYYY-MM-DD format for archive folders
5. **Update this index** - Keep this README current when adding docs
