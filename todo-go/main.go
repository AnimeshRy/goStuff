package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/mergestat/timediff"
	"github.com/spf13/cobra"
)

type Task struct {
	ID          int
	Description string
	CreatedAt   time.Time
	IsCompleted bool
}

var dataFile string

func init() {
	homeDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting home directory:", err)
		os.Exit(1)
	}
	dataFile = filepath.Join(homeDir, ".tasks.csv")
}

func loadFile(filepath string) (*os.File, error) {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading")
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, err
	}

	return f, nil
}

func closeFile(f *os.File) error {
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return f.Close()
}

func loadTasks() ([]Task, error) {
	f, err := loadFile(dataFile)
	if err != nil {
		return nil, err
	}
	defer closeFile(f)

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()

	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	var tasks []Task
	for i, record := range records {
		if i == 0 {
			continue // Skip Headers
		}
		id, _ := strconv.Atoi(record[0])
		createdAt, _ := time.Parse(time.RFC3339, record[2])
		isComplete, _ := strconv.ParseBool(record[3])

		tasks = append(tasks, Task{
			ID:          id,
			Description: record[1],
			CreatedAt:   createdAt,
			IsCompleted: isComplete,
		})
	}

	return tasks, nil
}

func saveTasks(tasks []Task) error {
	f, err := loadFile(dataFile)
	if err != nil {
		return err
	}
	defer closeFile(f)

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write Headers
	err = writer.Write([]string{"ID", "Description", "CreatedAt", "IsCompleted"})
	if err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write Tasks
	for _, task := range tasks {
		err := writer.Write([]string{
			strconv.Itoa(task.ID),
			task.Description,
			task.CreatedAt.Format(time.RFC3339),
			strconv.FormatBool(task.IsCompleted),
		})
		if err != nil {
			return fmt.Errorf("failed to write task to CSV: %w", err)
		}
	}
	return nil
}

func getNextID(tasks []Task) int {
	maxID := 0
	for _, task := range tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}
	return maxID + 1
}

var rootCmd = &cobra.Command{
	Use:   "tasks",
	Short: "A simple CLI todo application",
}

var addCmd = &cobra.Command{
	Use:   "add [description]",
	Short: "Add a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := loadTasks()
		if err != nil {
			return err
		}

		newTask := Task{
			ID:          getNextID(tasks),
			Description: args[0],
			CreatedAt:   time.Now(),
			IsCompleted: false,
		}

		tasks = append(tasks, newTask)
		if err := saveTasks(tasks); err != nil {
			return err
		}

		fmt.Printf("Added tasks %d: %s\n", newTask.ID, newTask.Description)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		showAll, _ := cmd.Flags().GetBool("all")
		tasks, err := loadTasks()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		if showAll {
			fmt.Fprintln(w, "ID\tTask\tCreated\tDone")
		} else {
			fmt.Fprintln(w, "ID\tTask\tCreated")
		}

		for _, task := range tasks {
			if !showAll && task.IsCompleted {
				continue
			}

			timeAgo := timediff.TimeDiff(task.CreatedAt)
			if showAll {
				fmt.Fprintf(w, "%d\t%s\t%s\t%v\n",
					task.ID, task.Description, timeAgo, task.IsCompleted)
			} else {
				fmt.Fprintf(w, "%d\t%s\t%s\n",
					task.ID, task.Description, timeAgo)
			}
		}
		return w.Flush()
	},
}

var completeCmd = &cobra.Command{
	Use:   "complete [taskID]",
	Short: "Mark a task as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
		}

		tasks, err := loadTasks()
		if err != nil {
			return err
		}

		found := false
		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].IsCompleted = true
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("task with ID %d not found", id)
		}

		if err := saveTasks(tasks); err != nil {
			return err
		}

		fmt.Printf("Marked task %d as complete\n", id)
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [taskID]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
		}

		tasks, err := loadTasks()
		if err != nil {
			return err
		}

		found := false
		newTasks := make([]Task, 0, len(tasks)-1)
		for _, task := range tasks {
			if task.ID == id {
				found = true
				continue
			}
			newTasks = append(newTasks, task)
		}

		if !found {
			return fmt.Errorf("task with ID %d not found", id)
		}

		if err := saveTasks(newTasks); err != nil {
			return err
		}

		fmt.Printf("Deleted task %d\n", id)
		return nil
	},
}

func main() {
	listCmd.Flags().BoolP("all", "a", false, "Show all tasks including completed ones")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
