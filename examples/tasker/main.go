package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	gstring "github.com/nurysso/gstrings"
)

const dataFile = "tasks.json"

func main() {
	store, err := LoadStore(dataFile)
	if err != nil {
		printError("Could not load {file}: {err}", gstring.With("file", dataFile, "err", err))
		os.Exit(1)
	}

	args := os.Args[1:]

	if len(args) == 0 {
		printBanner()
		printHelp()
		return
	}

	cmd := args[0]

	switch cmd {

	case "list":
		printBanner()
		tasks := store.Tasks
		header := "all tasks"

		for _, a := range args[1:] {
			if strings.HasPrefix(a, "--status=") {
				st := Status(strings.TrimPrefix(a, "--status="))
				tasks = store.FilterByStatus(st)
				header = gstring.Sprintf("status: {s}", gstring.With("s", string(st)))
			} else if strings.HasPrefix(a, "--project=") {
				proj := strings.TrimPrefix(a, "--project=")
				tasks = store.FilterByProject(proj)
				header = gstring.Sprintf("project: {p}", gstring.With("p", proj))
			}
		}

		printTaskList(tasks, header)
		printSummary(store)

	case "add":
		printBanner()
		reader := bufio.NewReader(os.Stdin)

		title := ""
		if len(args) > 1 {
			title = strings.Join(args[1:], " ")
		} else {
			title = prompt(reader, "Title", "")
		}
		if title == "" {
			printError("Title cannot be empty.", gstring.With())
			os.Exit(1)
		}

		project := prompt(reader, "Project", "General")
		priInput := prompt(reader, "Priority [high/medium/low]", "medium")
		dueDate := prompt(reader, "Due date [YYYY-MM-DD or blank]", "")
		notes := prompt(reader, "Notes [optional]", "")

		var pri Priority
		switch strings.ToLower(priInput) {
		case "high", "h":
			pri = PriorityHigh
		case "low", "l":
			pri = PriorityLow
		default:
			pri = PriorityMedium
		}

		t := store.Add(gstring.Strip(title), gstring.Strip(project), pri, gstring.Strip(dueDate), gstring.Strip(notes))
		if err := store.Save(); err != nil {
			printError("Failed to save: {err}", gstring.With("err", err))
			os.Exit(1)
		}
		fmt.Println()
		printSuccess(
			"Task #{id} created: {title}",
			gstring.With("id", t.ID, "title", t.Title),
		)
		fmt.Println()
		printTaskDetail(t)

	case "show":
		printBanner()
		id := requireID(args)
		t := store.Find(id)
		if t == nil {
			printError("No task with ID #{id}", gstring.With("id", id))
			os.Exit(1)
		}
		printTaskDetail(t)

	case "done":
		id := requireID(args)
		t := store.Find(id)
		if t == nil {
			printError("No task with ID #{id}", gstring.With("id", id))
			os.Exit(1)
		}
		t.Status = StatusDone
		store.Save()
		printSuccess(
			"Task #{id} marked as done: {title}",
			gstring.With("id", t.ID, "title", t.Title),
		)

	case "start":
		id := requireID(args)
		t := store.Find(id)
		if t == nil {
			printError("No task with ID #{id}", gstring.With("id", id))
			os.Exit(1)
		}
		t.Status = StatusInProgress
		store.Save()
		printSuccess(
			"Task #{id} started: {title}",
			gstring.With("id", t.ID, "title", t.Title),
		)

	case "delete", "rm":
		id := requireID(args)
		t := store.Find(id)
		if t == nil {
			printError("No task with ID #{id}", gstring.With("id", id))
			os.Exit(1)
		}
		title := t.Title
		store.Delete(id)
		store.Save()
		printSuccess(
			"Deleted task #{id}: {title}",
			gstring.With("id", id, "title", title),
		)

	case "summary", "stats":
		printBanner()
		printSummary(store)

	case "demo":
		// Use a fresh in-memory store — don't touch the real data file
		demoStore := &Store{file: "", NextID: 1}
		runDemo(demoStore)

	case "help", "--help", "-h":
		printBanner()
		printHelp()

	default:
		printBanner()
		printError(
			"Unknown command: {cmd}. Run 'tasker help' for usage.",
			gstring.With("cmd", cmd),
		)
		os.Exit(1)
	}
}

//  Helpers

func requireID(args []string) int {
	if len(args) < 2 {
		printError("Please provide a task ID.", gstring.With())
		os.Exit(1)
	}
	id, err := strconv.Atoi(args[1])
	if err != nil {
		printError("Invalid ID: {v}", gstring.With("v", args[1]))
		os.Exit(1)
	}
	return id
}

func prompt(r *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf(string(gstring.ColorGray)+"  %s [%s]: "+string(gstring.ColorNone), label, def)
	} else {
		fmt.Printf(string(gstring.ColorGray)+"  %s: "+string(gstring.ColorNone), label)
	}
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}
