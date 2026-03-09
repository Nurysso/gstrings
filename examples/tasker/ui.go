package main

import (
	"fmt"
	"strings"

	gstring "github.com/nurysso/gstrings"
)

// Helpers

func priorityColor(p Priority) gstring.Color {
	switch p {
	case PriorityHigh:
		return gstring.ColorRed
	case PriorityMedium:
		return gstring.ColorYellow
	case PriorityLow:
		return gstring.ColorGreen
	}
	return gstring.ColorNone
}

func statusColor(s Status) gstring.Color {
	switch s {
	case StatusDone:
		return gstring.ColorGreen
	case StatusInProgress:
		return gstring.ColorCyan
	case StatusTodo:
		return gstring.ColorGray
	}
	return gstring.ColorNone
}

func priorityIcon(p Priority) string {
	switch p {
	case PriorityHigh:
		return "!!!"
	case PriorityMedium:
		return " ! "
	case PriorityLow:
		return " · "
	}
	return "   "
}

func statusIcon(s Status) string {
	switch s {
	case StatusDone:
		return "[✓]"
	case StatusInProgress:
		return "[~]"
	case StatusTodo:
		return "[ ]"
	}
	return "   "
}

// Header / Banner

func printBanner() {
	width := 52
	// Define a reset constant for readability if not in your package
	const reset = "\033[0m"

	fmt.Println()

	// Top Border
	fmt.Println(string(gstring.ColorCyan) + gstring.Repeat("─", width) + reset)

	// Title (Bold)
	title := gstring.Center("✦  TASKER  ✦", width, ' ')
	fmt.Println(string(gstring.ColorBold) + title + reset)

	// Subtitle (Gray)
	subtitle := gstring.Center("CLI task manager · powered by gstring", width, ' ')
	fmt.Println(string(gstring.ColorGray) + subtitle + reset)

	// Bottom Border
	fmt.Println(string(gstring.ColorCyan) + gstring.Repeat("─", width) + reset)

	fmt.Println()
}

func printSection(title string) {
	bar := gstring.Repeat("─", 52)
	label := gstring.Center(" "+strings.ToUpper(title)+" ", 52, '─')
	_ = bar
	fmt.Println()
	fmt.Println(string(gstring.ColorBlue) + label + string(gstring.ColorNone))
}

//  Task List

func printTaskList(tasks []*Task, header string) {
	printSection(header)
	if len(tasks) == 0 {
		fmt.Println(string(gstring.ColorGray) + "  (no tasks)" + string(gstring.ColorNone))
		fmt.Println()
		return
	}

	// Build rows for AutoWidth
	rows := make([][]any, len(tasks))
	for i, t := range tasks {
		rows[i] = []any{
			fmt.Sprintf("#%d", t.ID),
			statusIcon(t.Status) + " " + priorityIcon(t.Priority),
			gstring.Truncate(t.Title, 32),
			t.Project,
			string(t.Priority),
			dueDateDisplay(t.DueDate),
		}
	}

	tbl := gstring.NewTable().
		Col("ID", 4, gstring.AlignRight).
		Col("S P", 7, gstring.AlignCenter).
		Col("TITLE", 32, gstring.AlignLeft).
		Col("PROJECT", 12, gstring.AlignLeft).
		Col("PRIORITY", 8, gstring.AlignLeft).
		Col("DUE", 10, gstring.AlignLeft).
		HeaderColor(gstring.ColorCyan).
		Separator("  ")

	tbl.AutoWidth(rows)
	tbl.Header()

	for i, t := range tasks {
		statusC := statusColor(t.Status)
		priC := priorityColor(t.Priority)

		idStr := fmt.Sprintf("#%d", t.ID)
		badge := statusIcon(t.Status) + " " + priorityIcon(t.Priority)
		title := gstring.Truncate(t.Title, 32)
		project := t.Project
		priority := string(t.Priority)
		due := dueDateDisplay(t.DueDate)

		// Color individual cells manually since Row() colors the whole cell
		_ = rows[i]
		fmt.Print(
			string(gstring.ColorGray)+gstring.Pad(idStr, 4, ' ')+"  "+string(gstring.ColorNone),
			string(statusC)+gstring.Center(badge, 7, ' ')+"  "+string(gstring.ColorNone),
			gstring.Pad(title, 32, ' ')+"  ",
			string(gstring.ColorBlue)+gstring.Pad(project, 12, ' ')+"  "+string(gstring.ColorNone),
			string(priC)+gstring.Pad(priority, 8, ' ')+"  "+string(gstring.ColorNone),
			dueDateColored(due),
			"\n",
		)
	}
	fmt.Println()
}

func dueDateDisplay(d string) string {
	if d == "" {
		return "—"
	}
	return d
}

func dueDateColored(d string) string {
	if d == "—" {
		return string(gstring.ColorGray) + d + string(gstring.ColorNone)
	}
	return string(gstring.ColorYellow) + d + string(gstring.ColorNone)
}

// Task Detail

func printTaskDetail(t *Task) {
	printSection(gstring.Sprintf("task #{id}", gstring.With("id", t.ID)))
	fmt.Println()

	width := 48

	gstring.Println(
		"  {icon} {title}",
		gstring.With("icon", statusIcon(t.Status), "title", t.Title),
	)
	fmt.Println()

	// Info rows using Row builder
	printDetailRow := func(label, value string, valueColor gstring.Color) {
		fmt.Print(
			string(gstring.ColorGray)+gstring.Pad(label, 12, ' ')+string(gstring.ColorNone),
			"  ",
			string(valueColor)+value+string(gstring.ColorNone),
			"\n",
		)
	}

	printDetailRow("ID", fmt.Sprintf("#%d", t.ID), gstring.ColorWhite)
	printDetailRow("Status", string(t.Status), statusColor(t.Status))
	printDetailRow("Priority", string(t.Priority), priorityColor(t.Priority))
	printDetailRow("Project", t.Project, gstring.ColorBlue)
	printDetailRow("Due Date", dueDateDisplay(t.DueDate), gstring.ColorYellow)
	printDetailRow("Created", t.CreatedAt, gstring.ColorGray)

	if t.Notes != "" {
		fmt.Println()
		fmt.Println(string(gstring.ColorGray) + "  Notes:" + string(gstring.ColorNone))
		wrapped := gstring.Wrap(t.Notes, width-4)
		for _, line := range strings.Split(wrapped, "\n") {
			fmt.Println("    " + line)
		}
	}

	fmt.Println()
	fmt.Println(string(gstring.ColorCyan) + gstring.Repeat("─", width) + string(gstring.ColorNone))
	fmt.Println()
}

//  Summary / Stats

func printSummary(s *Store) {
	printSection("summary")
	todo, inProg, done := s.Stats()
	total := todo + inProg + done

	fmt.Println()

	// Stats row using Row builder
	gstring.NewRow().
		Left(string(gstring.ColorGray)+"Total"+string(gstring.ColorNone), 10).
		Sep("·").
		Left(string(gstring.ColorWhite)+fmt.Sprintf("%d tasks", total)+string(gstring.ColorNone), 12).
		Print()

	gstring.NewRow().
		Left(string(gstring.ColorGray)+"Todo"+string(gstring.ColorNone), 10).
		Sep("·").
		Left(string(gstring.ColorGray)+fmt.Sprintf("%d", todo)+string(gstring.ColorNone), 12).
		Print()

	gstring.NewRow().
		Left(string(gstring.ColorGray)+"In Progress"+string(gstring.ColorNone), 10).
		Sep("·").
		Left(string(gstring.ColorCyan)+fmt.Sprintf("%d", inProg)+string(gstring.ColorNone), 12).
		Print()

	gstring.NewRow().
		Left(string(gstring.ColorGray)+"Done"+string(gstring.ColorNone), 10).
		Sep("·").
		Left(string(gstring.ColorGreen)+fmt.Sprintf("%d", done)+string(gstring.ColorNone), 12).
		Print()

	fmt.Println()

	// Progress bar
	if total > 0 {
		barWidth := 30
		filled := (done * barWidth) / total
		bar := string(gstring.ColorGreen) +
			gstring.Repeat("█", filled) +
			string(gstring.ColorGray) +
			gstring.Repeat("░", barWidth-filled) +
			string(gstring.ColorNone)

		pct := (done * 100) / total
		gstring.Println(
			"  Progress  [{bar}] {pct}% complete",
			gstring.With("bar", bar, "pct", pct),
		)
		fmt.Println()
	}

	// Projects breakdown
	projects := s.Projects()
	if len(projects) > 0 {
		fmt.Println(string(gstring.ColorGray) + "  Projects:" + string(gstring.ColorNone))
		for _, p := range projects {
			tasks := s.FilterByProject(p)
			doneCount := 0
			for _, t := range tasks {
				if t.Status == StatusDone {
					doneCount++
				}
			}
			gstring.Println(
				"    {project}  {done}/{total} done",
				gstring.With(
					"project", string(gstring.ColorBlue)+gstring.Pad(p, 16, ' ')+string(gstring.ColorNone),
					"done", doneCount,
					"total", len(tasks),
				),
			)
		}
		fmt.Println()
	}
}

// Success / Error messages

func printSuccess(template string, vars gstring.Vars) {
	msg := gstring.Sprintf(template, vars)
	fmt.Println(string(gstring.ColorGreen) + "  ✓ " + msg + string(gstring.ColorNone))
}

func printError(template string, vars gstring.Vars) {
	msg := gstring.Sprintf(template, vars)
	fmt.Println(string(gstring.ColorRed) + "  ✗ " + msg + string(gstring.ColorNone))
}

func printInfo(template string, vars gstring.Vars) {
	msg := gstring.Sprintf(template, vars)
	fmt.Println(string(gstring.ColorCyan) + "  · " + msg + string(gstring.ColorNone))
}

//  Help

func printHelp() {
	printSection("commands")
	fmt.Println()

	cmds := []struct{ cmd, args, desc string }{
		{"list", "", "List all tasks"},
		{"list", "--status=<s>", "Filter by status (todo, in-progress, done)"},
		{"list", "--project=<p>", "Filter by project name"},
		{"add", "<title>", "Add a task (interactive prompts follow)"},
		{"show", "<id>", "Show full task details"},
		{"done", "<id>", "Mark task as done"},
		{"start", "<id>", "Mark task as in-progress"},
		{"delete", "<id>", "Delete a task"},
		{"summary", "", "Show progress summary"},
		{"demo", "", "Load demo data and show all features"},
		{"help", "", "Show this help"},
	}

	tbl := gstring.NewTable().
		Col("COMMAND", 8, gstring.AlignLeft).
		Col("ARGS", 20, gstring.AlignLeft).
		Col("DESCRIPTION", 36, gstring.AlignLeft).
		HeaderColor(gstring.ColorCyan).
		Separator("  ")

	tbl.Header()
	for _, c := range cmds {
		tbl.Row(c.cmd, c.args, c.desc)
	}
	fmt.Println()
}

//  Demo seed

func runDemo(s *Store) {
	printBanner()
	printInfo("Loading demo data...", gstring.With())
	fmt.Println()

	// Seed tasks
	tasks := []struct {
		title, project string
		priority       Priority
		due, notes     string
		status         Status
	}{
		{"Design landing page", "Website", PriorityHigh, "2025-04-10",
			"Include hero section with CTA button and responsive layout for mobile.", StatusInProgress},
		{"Write unit tests for API", "Backend", PriorityHigh, "2025-04-05",
			"Cover auth endpoints, error cases, and rate limiting logic.", StatusTodo},
		{"Set up CI/CD pipeline", "Backend", PriorityMedium, "2025-04-12",
			"GitHub Actions preferred. Deploy to staging on merge to main.", StatusTodo},
		{"Migrate database schema", "Backend", PriorityHigh, "2025-04-08", "", StatusDone},
		{"Write onboarding copy", "Website", PriorityLow, "2025-04-20",
			"Keep tone friendly. Max 3 steps. Use short sentences.", StatusTodo},
		{"Performance audit", "Website", PriorityMedium, "", "", StatusTodo},
		{"Fix login redirect bug", "Backend", PriorityHigh, "2025-04-03", "", StatusDone},
		{"Add dark mode toggle", "Website", PriorityLow, "2025-04-25",
			"Persist preference in localStorage. Follow system default initially.", StatusInProgress},
	}

	for _, td := range tasks {
		t := s.Add(td.title, td.project, td.priority, td.due, td.notes)
		t.Status = td.status
	}

	// Show everything
	printTaskList(s.Tasks, "all tasks")

	// Detail view for one
	printTaskDetail(s.Tasks[0])

	// Filtered view
	inProg := s.FilterByStatus(StatusInProgress)
	printTaskList(inProg, "in progress")

	// Summary
	printSummary(s)
}
