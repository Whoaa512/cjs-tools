package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

type Link struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type LinksDatabase struct {
	Links []Link `json:"links"`
}

var dbFile string
var rootCmd = &cobra.Command{Use: "quicklinks"}
var db LinksDatabase

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error finding home directory: %v", err)
	}
	dbFile = filepath.Join(home, ".config", "quicklinks", "db.json")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(goCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(startDaemonCmd)
	loadLinks()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var addCmd = &cobra.Command{
	Use:   "add [name] [url]",
	Short: "Add a new link",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		url := args[1]
		description, _ := cmd.Flags().GetString("description")

		link := Link{
			Name:        name,
			URL:         url,
			Description: description,
		}

		db.Links = append(db.Links, link)
		saveLinks()
		fmt.Printf("Added link: %s -> %s\n", name, url)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all links",
	Run: func(cmd *cobra.Command, args []string) {
		for _, link := range db.Links {
			fmt.Printf("%s: %s (%s)\n", link.Name, link.URL, link.Description)
		}
	},
}

var goCmd = &cobra.Command{
	Use:   "go [name]",
	Short: "Open a link in the default browser",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		for _, link := range db.Links {
			if link.Name == name {
				openBrowser(link.URL)
				return
			}
		}
		fmt.Printf("Link not found: %s\n", name)
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a link",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		for i, link := range db.Links {
			if link.Name == name {
				db.Links = append(db.Links[:i], db.Links[i+1:]...)
				saveLinks()
				fmt.Printf("Removed link: %s\n", name)
				return
			}
		}
		fmt.Printf("Link not found: %s\n", name)
	},
}

var startDaemonCmd = &cobra.Command{
	Use:   "start-daemon",
	Short: "Start a daemon server",
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/links", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" {
				jsonData, err := json.Marshal(db)
				if err != nil {
					http.Error(
						w,
						"Error converting links data to JSON",
						http.StatusInternalServerError,
					)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonData)
			}
		})
		fmt.Println("Starting server at port 8080...")
		log.Fatal(http.ListenAndServe(":8080", nil))
	},
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func saveLinks() {
	jsonData, err := json.Marshal(db)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}
	err = ioutil.WriteFile(dbFile, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}
}

func loadLinks() {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return // File does not exist
	}
	jsonData, err := ioutil.ReadFile(dbFile)
	if err != nil {
		log.Fatalf("Error reading from file: %v", err)
	}
	err = json.Unmarshal(jsonData, &db)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}
}

func init() {
	addCmd.Flags().StringP("description", "d", "", "Description of the link")
}
