# r8s v1.3.2 Release Notes

**Release Date:** March 5, 2026
**Focus:** Demo Readiness & Team Enablement

This release prepares r8s for wider team adoption, focusing on documentation, demo capabilities, and streamlined contribution workflows.

## 🚀 Key Highlights

### 🎭 Demo Mode & Training Resources
We've added a comprehensive **Demo Kit** to help you present r8s to your team or customers:
- **`scripts/setup-demo-bundles.sh`**: Instantly generates two realistic support bundles with "Critical" and "Deep Dive" scenarios (Etcd Quorum Loss, CrashLoops, OOMs).
- **`docs/DEMO_SCRIPT.md`**: A minute-by-minute speaker script with "talking points" and anticipated Q&A.

### 📝 Contribution Guide
We've overhauled `CONTRIBUTING.md` to include:
- Standardized **Bug Report** and **Feature Request** templates.
- Clearer testing goals (80% coverage target).
- Branch workflow guidelines for sprint management.

### 🧹 Housekeeping
- Moved ad-hoc debug scripts to `scripts/` to keep the project root clean.
- Updated documentation to reflect the latest version v1.3.2.

## 📦 Usage

**Setup the demo:**
```bash
./scripts/setup-demo-bundles.sh
export r8sbundle1="./support-bundle-2026-03-01"
export r8sbundle2="./support-bundle-2026-03-02"
```

**Run the demo:**
```bash
r8s analyze $r8sbundle1
r8s ask $r8sbundle1 "what is the main issue?"
```

## 🔗 Links
- [Demo Script](docs/DEMO_SCRIPT.md)
- [Contributing Guidelines](CONTRIBUTING.md)
