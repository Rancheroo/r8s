# TUI Architecture and Entry Points

**Purpose:** Document how `internal/tui` connects to business logic for future wizard implementation  
**Date:** 2026-02-16  
**Status:** Sprint 6 Foundation (enables v0.7.3 wizard work)

---

## Overview

The TUI (Terminal User Interface) uses the **Bubble Tea** framework with a Model-Update-View architecture. This document maps the entry points for:
1. Bundle loading and initialization
2. Navigation flows
3. Data access patterns
4. Extension points for wizard UI (v0.7.3)

---

## Architecture Layers

```
┌─────────────────────────────────────────┐
│  View Layer (Bubble Tea Views)          │
│  - app.go: Main application view        │
│  - helpers.go: View helpers             │
│  - pods.go, events.go: Specific views   │
└──────────────────┬──────────────────────┘
                   │ Update() Msgs
┌──────────────────▼──────────────────────┐
│  Controller Layer (State Management)    │
│  - app.go: Model struct, state machine  │
│  - update.go: Message handlers          │
└──────────────────┬──────────────────────┘
                   │ Fetch/Data calls
┌──────────────────▼──────────────────────┐
│  Service Layer (Business Logic)         │
│  - fetch.go: Data fetching              │
│  - internal/datasource: Data access     │
│  - internal/bundle: Bundle parsing      │
└─────────────────────────────────────────┘
```

---

## Key Entry Points

### 1. Application Initialization

**File:** `internal/tui/app.go`  
**Function:** `InitialModel(bundlePath string)`

```go
// Entry point for TUI startup
func InitialModel(bundlePath string) Model {
    model := Model{
        bundlePath: bundlePath,
        // Initial state setup
    }
    // Triggers async bundle loading
    return model
}
```

**Extension for Wizard:** Wizard can inject pre-load questions here.

---

### 2. Bundle Loading Flow

**File:** `internal/tui/fetch.go`  
**Function:** `LoadBundle`

```go
// Async bundle loading triggered by Init()
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.LoadBundle(),           // ← Entry Point 1
        m.ListenForUpdates(),     // ← Entry Point 2
    )
}

func (m Model) LoadBundle() tea.Cmd {
    return func() tea.Msg {
        // Calls internal/bundle.LoadBundle()
        bundle, err := bundle.Load(m.bundlePath)
        return BundleLoadedMsg{bundle, err}
    }
}
```

**Wizard Extension:** Can intercept before/after loading to show progress.

---

### 3. Navigation State Machine

**File:** `internal/tui/app.go`  
**Type:** `Model.state` (enum)

```go
type state int

const (
    loadingState state = iota      // Initial bundle load
    dashboardState                 // Main dashboard view
    podListState                   // Pod listing
    eventListState                 // Event listing
    logViewState                   // Log viewer
    podDescribeState               // Pod describe view
    // ← Wizard State Extension Point: wizardState
)

type Model struct {
    state        state        // ← Current navigation state
    previousState state       // ← For "back" navigation
    
    bundle       *bundle.Bundle  // ← Data source
    dataSource   datasource.DataSource
    
    // View-specific state
    selectedPod     string
    selectedNamespace string
    logFilter       string
}
```

**Wizard Extension:** Add `wizardState` with step tracking.

---

### 4. Update Message Flow

**File:** `internal/tui/app.go`  
**Function:** `Update(msg tea.Msg)`

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyPress(msg)  // ← Entry Point
        
    case BundleLoadedMsg:
        return m.handleBundleLoaded(msg)  // ← Entry Point
        
    case ErrorMsg:
        return m.handleError(msg)  // ← Entry Point
        
    // ← Wizard Message Extension Point
    case WizardNextStepMsg:
        return m.advanceWizard(msg)
    }
    
    return m, nil
}
```

**Key Handlers:**
- `handleKeyPress`: Keyboard navigation (arrows, enter, esc, etc.)
- `handleBundleLoaded`: Transition from loading → dashboard
- `handleError`: Error display and recovery

---

### 5. View Rendering

**File:** `internal/tui/app.go`  
**Function:** `View() string`

```go
func (m Model) View() string {
    switch m.state {
    case loadingState:
        return m.loadingView()     // ← Spinner/progress
        
    case dashboardState:
        return m.dashboardView()   // ← Main dashboard
        
    case podListState:
        return m.podListView()     // ← Pod selection
        
    case logViewState:
        return m.logView()         // ← Log viewer
        
    // ← Wizard View Extension Point
    case wizardState:
        return m.wizardView()
    }
    
    return "Unknown state"
}
```

---

### 6. Data Access Patterns

**File:** `internal/tui/fetch.go`  
**Pattern:** Async fetching with message passing

```go
// Generic data fetch pattern
type DataFetchedMsg struct {
    Data interface{}
    Err  error
}

func fetchData(dataSource datasource.DataSource) tea.Cmd {
    return func() tea.Msg {
        data, err := dataSource.GetData()
        return DataFetchedMsg{data, err}
    }
}
```

**Data Sources:**
- `BundleDataSource`: Live bundle parsing
- `EmbeddedDataSource`: Demo/embedded data

**Wizard Reuse:** Wizard can use same pattern for step data.

---

### 7. Component Library Location

**Existing Components:**
- `internal/tui/components/table.go`: Pod/event tables
- `internal/tui/components/panel.go`: Info panels
- `internal/tui/components/logviewer.go`: Log display

**Wizard Components Needed (v0.7.3):**
- `internal/tui/components/wizard.go`: Step container
- `internal/tui/components/question.go`: Interactive prompts
- `internal/tui/components/progress.go`: Progress indicators

---

## Extension Points for Wizard (v0.7.3)

### Adding Wizard State

```go
// In app.go, add to state enum
const (
    // ... existing states ...
    wizardState state = iota
)

// Add to Model struct
type Model struct {
    // ... existing fields ...
    wizardStep   int
    wizardData   WizardData
}

type WizardData struct {
    IssueType    string
    Component    string
    SuspectedCause string
}
```

### Wizard Message Types

```go
// New message types for wizard
type WizardNextStepMsg struct{}
type WizardPrevStepMsg struct{}
type WizardSelectOptionMsg struct {
    Option string
}
type WizardCompleteMsg struct {
    Diagnosis DiagnosisResult
}
```

### Wizard View Structure

```go
func (m Model) wizardView() string {
    switch m.wizardStep {
    case 1:
        return m.wizardIssueTypeView()
    case 2:
        return m.wizardComponentView()
    case 3:
        return m.wizardAnalysisView()
    case 4:
        return m.wizardResultView()
    }
    return ""
}
```

---

## Testing Entry Points

### Interface Extraction (Sprint 6)

To enable unit testing without Bubble Tea runtime:

```go
// internal/tui/interfaces.go

// Controller interface for testing
type Controller interface {
    LoadBundle(path string) error
    GetState() state
    SetState(s state)
    GetBundle() *bundle.Bundle
    SelectPod(name string)
    GetSelectedPod() string
}

// Concrete implementation
type TUIController struct {
    model *Model
}

func (c *TUIController) LoadBundle(path string) error {
    // Implementation
}

// Test can use mock:
type MockController struct {
    mock.Mock
}
```

---

## Keyboard Shortcuts Reference

| Key | Action | State |
|-----|--------|-------|
| `↑/↓` or `j/k` | Navigate | All list views |
| `Enter` | Drill down / Select | All |
| `Esc` or `b` | Back | All detail views |
| `q` | Quit | All |
| `d` | Describe (JSON) | Pod/event views |
| `r` | Refresh | Dashboard |
| `m` | Toggle dashboard | Dashboard |
| `c` | Classic/cluster view | Dashboard |
| `/` | Search | Log view |
| `?` | Help | All |

**Wizard Shortcuts (v0.7.3):**
- `n` / `→` : Next step
- `p` / `←` : Previous step
- `Enter` : Confirm selection

---

## Dependencies

**External:**
- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/lipgloss`: Styling
- `github.com/charmbracelet/bubbles`: Components

**Internal:**
- `internal/bundle`: Bundle parsing
- `internal/datasource`: Data abstraction
- `internal/rancher`: Rancher types

---

## Sprint 6 Actions

1. ✅ Document entry points (this file)
2. 🔴 Extract `Controller` interface for testing
3. 🔴 Add `state` getter/setter for external control
4. 🔴 Document existing component library

---

## v0.7.3 Wizard Preparation

**Pre-work Checklist:**
- [ ] Controller interface extracted and tested
- [ ] Component library documented
- [ ] State machine supports extension
- [ ] Message types defined for wizard

**Wizard Implementation Order:**
1. Add `wizardState` to enum
2. Create `internal/tui/wizard` package
3. Implement step components
4. Add keyboard shortcuts
5. Integrate with pattern engine

---

*Document enables TUI-Expert role to contribute effectively in v0.7.3.*
