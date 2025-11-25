// Test program to demonstrate the describe feature with mock pod data
// This bypasses the need for a real Rancher connection
//
// Usage: go run test_describe.go
//
// Test scenarios:
// 1. Navigate pods with j/k keys
// 2. Press 'd' to describe selected pod
// 3. Verify JSON display in describe view
// 4. Test exit methods: 'Esc', 'q', 'd' from describe view
// 5. Press '?' for help

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

// Simplified types for standalone testing
type Pod struct {
	Name        string
	Namespace   string
	State       string
	NodeName    string
}

type testApp struct {
	width           int
	height          int
	pods            []Pod
	table           table.Model
	showHelp        bool
	showingDescribe bool
	describeTitle   string
	describeContent string
}

func newTestApp() *testApp {
	// Create mock pod data
	pods := []Pod{
		{Name: "nginx-deployment-7d6c9f8c5d-abc12", Namespace: "default", State: "Running", NodeName: "worker-1"},
		{Name: "nginx-deployment-7d6c9f8c5d-xyz89", Namespace: "default", State: "Running", NodeName: "worker-2"},
		{Name: "redis-master-0", Namespace: "default", State: "Running", NodeName: "worker-1"},
		{Name: "postgres-statefulset-0", Namespace: "production", State: "Running", NodeName: "worker-3"},
		{Name: "api-server-6b8d9c7f5-mnp45", Namespace: "production", State: "Running", NodeName: "worker-2"},
	}

	return &testApp{
		pods: pods,
	}
}

func (a *testApp) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (a *testApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle help screen
		if a.showHelp {
			if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
				a.showHelp = false
			}
			return a, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if a.showingDescribe {
				// Exit describe view
				a.showingDescribe = false
				a.describeContent = ""
				a.describeTitle = ""
				return a, nil
			}
			return a, tea.Quit
		case "?":
			a.showHelp = true
			return a, nil
		case "esc":
			if a.showingDescribe {
				a.showingDescribe = false
				a.describeContent = ""
				a.describeTitle = ""
			}
			return a, nil
		case "d":
			if a.showingDescribe {
				// Exit describe view
				a.showingDescribe = false
				a.describeContent = ""
				a.describeTitle = ""
				return a, nil
			}
			// Enter describe view
			return a, a.handleDescribe()
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateTable()
	}

	// Update table
	var cmd tea.Cmd
	a.table, cmd = a.table.Update(msg)
	return a, cmd
}

func (a *testApp) View() string {
	if a.showHelp {
		return renderHelp()
	}

	if a.showingDescribe {
		return a.renderDescribeView()
	}

	// Build view components
	colorCyan := lipgloss.Color("#00CED1")
	breadcrumbStyle := lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))

	breadcrumb := breadcrumbStyle.Render("Test Cluster > default namespace > Pods")
	statusText := fmt.Sprintf(" %d pods | Press 'd' to describe selected pod | '?' for help | 'q' to quit ", len(a.pods))
	status := statusStyle.Render(statusText)

	// Render table
	tableView := a.table.View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		breadcrumb,
		"",
		tableView,
		"",
		status,
	)
}

func (a *testApp) updateTable() {
	if len(a.pods) == 0 {
		return
	}

	colorCyan := lipgloss.Color("#00CED1")
	headerStyle := lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	baseStyle := lipgloss.NewStyle()

	columns := []table.Column{
		table.NewColumn("name", "NAME", 40),
		table.NewColumn("namespace", "NAMESPACE", 20),
		table.NewColumn("state", "STATE", 15),
		table.NewColumn("node", "NODE", 15),
	}

	rows := []table.Row{}
	for _, pod := range a.pods {
		rows = append(rows, table.NewRow(table.RowData{
			"name":      pod.Name,
			"namespace": pod.Namespace,
			"state":     pod.State,
			"node":      pod.NodeName,
		}))
	}

	pageSize := a.height - 8
	if pageSize < 5 {
		pageSize = 5
	}

	a.table = table.New(columns).
		WithRows(rows).
		HeaderStyle(headerStyle).
		WithBaseStyle(baseStyle).
		WithPageSize(pageSize).
		Focused(true).
		BorderRounded()
}

func (a *testApp) handleDescribe() tea.Cmd {
	if a.table.HighlightedRow().Data == nil {
		return nil
	}

	selected := a.table.HighlightedRow().Data
	podName := selected["name"].(string)
	namespace := selected["namespace"].(string)
	state := selected["state"].(string)
	node := selected["node"].(string)

	// Create mock pod details
	mockDetails := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      podName,
			"namespace": namespace,
			"labels": map[string]string{
				"app": strings.Split(podName, "-")[0],
			},
			"annotations": map[string]string{
				"description": "Mock pod data for describe feature testing",
			},
		},
		"spec": map[string]interface{}{
			"nodeName": node,
			"containers": []interface{}{
				map[string]interface{}{
					"name":  "app",
					"image": "nginx:latest",
					"ports": []interface{}{
						map[string]interface{}{
							"containerPort": 80,
							"protocol":      "TCP",
						},
					},
					"resources": map[string]interface{}{
						"requests": map[string]string{
							"cpu":    "100m",
							"memory": "128Mi",
						},
						"limits": map[string]string{
							"cpu":    "500m",
							"memory": "512Mi",
						},
					},
				},
			},
		},
		"status": map[string]interface{}{
			"phase":  state,
			"podIP":  "10.42.0." + fmt.Sprintf("%d", len(podName)%255),
			"hostIP": "192.168.1." + fmt.Sprintf("%d", len(node)%255),
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Ready",
					"status": "True",
				},
				map[string]interface{}{
					"type":   "Initialized",
					"status": "True",
				},
			},
		},
	}

	jsonBytes, _ := json.MarshalIndent(mockDetails, "", "  ")
	content := fmt.Sprintf("Pod Details (JSON):\n\n%s", string(jsonBytes))

	a.showingDescribe = true
	a.describeTitle = fmt.Sprintf("Pod: %s/%s", namespace, podName)
	a.describeContent = content

	return nil
}

func (a *testApp) renderDescribeView() string {
	colorCyan := lipgloss.Color("#00CED1")
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))

	titleBox := lipgloss.NewStyle().
		Foreground(colorCyan).
		Bold(true).
		Padding(0, 1).
		Render(fmt.Sprintf(" DESCRIBE: %s ", a.describeTitle))

	content := a.describeContent
	lines := strings.Split(content, "\n")
	maxLines := a.height - 8

	if len(lines) > maxLines {
		content = strings.Join(lines[:maxLines-1], "\n") + "\n... (truncated)"
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 2).
		Width(a.width - 4).
		Height(a.height - 6).
		Render(content)

	statusText := statusStyle.Render(" Press 'Esc', 'q' or 'd' to return | Scroll with mouse or arrow keys ")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleBox,
		"",
		contentBox,
		"",
		statusText,
	)
}

func renderHelp() string {
	colorCyan := lipgloss.Color("#00CED1")
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(2, 4).
		Width(80)

	helpText := `r9s - Describe Feature Test

KEYBINDINGS:
  j/k, ↑/↓     Navigate table rows
  d            Describe selected pod (opens describe view)
  Esc          Exit describe view
  q            Exit describe view / Quit app
  ?            Toggle this help screen

TEST SCENARIOS:
  1. Navigate pods using j/k or arrow keys
  2. Press 'd' on a selected pod to view details
  3. Verify JSON formatting in describe view
  4. Test exit methods: Esc, q, and d from describe view
  5. Verify cyan styling and borders
  6. Check title shows: "DESCRIBE: Pod: namespace/name"
  7. Verify status bar shows exit instructions

Press '?' or 'Esc' or 'q' to close this help`

	return helpStyle.Render(helpText)
}

func main() {
	fmt.Println("Starting r9s describe feature test...")
	fmt.Println("This test uses mock pod data to demonstrate the describe functionality.")
	fmt.Println("")

	app := newTestApp()
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
