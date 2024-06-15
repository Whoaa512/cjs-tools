package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
)

type Link struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type Database struct {
	Links []Link `json:"links"`
}

var (
	dbPath string
)

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	dbPath = filepath.Join(usr.HomeDir, ".config", "quicklinks", "db.json")
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "quicklinks",
		Short: "Quick Links Client",
	}

	var addCmd = &cobra.Command{
		Use:   "add [name] [url]",
		Short: "Add a new link",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				fmt.Println("Please provide a name and URL for the link")
				return
			}
			name := args[0]
			url := args[1]
			description, _ := cmd.Flags().GetString("description")
			addLink(name, url, description)
		},
	}
	addCmd.Flags().StringP("description", "d", "", "Description of the link")

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all links",
		Run: func(cmd *cobra.Command, args []string) {
			recent, _ := cmd.Flags().GetBool("recent")
			listLinks(recent)
		},
	}
	listCmd.Flags().BoolP("recent", "r", false, "List recent links")

	var goCmd = &cobra.Command{
		Use:   "go [name]",
		Short: "Open the link in the default browser",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Please provide the name of the link")
				return
			}
			name := args[0]
			openLink(name)
		},
	}

	var removeCmd = &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a link",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Please provide the name of the link to remove")
				return
			}
			name := args[0]
			removeLink(name)
		},
	}

	var daemonCmd = &cobra.Command{
		Use:   "start-daemon",
		Short: "Start the daemon server",
		Run: func(cmd *cobra.Command, args []string) {
			startDaemon()
		},
	}

	rootCmd.AddCommand(addCmd, listCmd, goCmd, removeCmd, daemonCmd)
	rootCmd.Execute()
}

func addLink(name, url, description string) {
	db := loadDatabase()
	link := Link{
		Name:        name,
		URL:         url,
		Description: description,
	}
	db.Links = append(db.Links, link)
	saveDatabase(db)
	fmt.Printf("Added link: %s\n", name)
}

func listLinks(recent bool) {
	db := loadDatabase()
	if recent {
		// List recent links (e.g., last 5)
		start := len(db.Links) - 5
		if start < 0 {
			start = 0
		}
		for _, link := range db.Links[start:] {
			fmt.Printf("%s - %s\n", link.Name, link.URL)
		}
	} else {
		for _, link := range db.Links {
			fmt.Printf("%s - %s\n", link.Name, link.URL)
		}
	}
}

func openLink(name string) {
	db := loadDatabase()
	for _, link := range db.Links {
		if link.Name == name {
			fmt.Printf("Opening link: %s\n", link.URL)
			// TODO: Open the link in the default browser
			return
		}
	}
	fmt.Printf("Link not found: %s\n", name)
}

func removeLink(name string) {
	db := loadDatabase()
	for i, link := range db.Links {
		if link.Name == name {
			db.Links = append(db.Links[:i], db.Links[i+1:]...)
			saveDatabase(db)
			fmt.Printf("Removed link: %s\n", name)
			return
		}
	}
	fmt.Printf("Link not found: %s\n", name)
}

func startDaemon() {
	fmt.Println("Starting daemon server...")
	// TODO: Implement the daemon server
}

func loadDatabase() Database {
	var db Database
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return db
	}
	json.Unmarshal(data, &db)
	return db
}

func saveDatabase(db Database) {
	data, _ := json.MarshalIndent(db, "", "  ")
	os.MkdirAll(filepath.Dir(dbPath), os.ModePerm)
	os.WriteFile(dbPath, data, 0644)
}
