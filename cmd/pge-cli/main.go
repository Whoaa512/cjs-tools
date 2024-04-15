package main

// CLI that takes in username, prompts for password then uses `playwright-go` to login to a website.
//
// Usage:
//   pge-cli [flags]
//   pge-cli [command]
//

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/charmbracelet/huh"
	"github.com/playwright-community/playwright-go"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "last-bill",
				Usage: "Login to PGE and get the last bill amount",
				Action: func(c *cli.Context) error {
					username := c.Args().First()
					if username == "" {
						fmt.Print("Username: ")
						fmt.Scanln(&username)
					}
					fmt.Print("Password: ")
					password, err := readPassword()
					if err != nil {
						return err
					}

					err = loginAndPrintLastBill(username, password)
					return err
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func readPassword() (string, error) {
	var password string
	err := huh.NewInput().
		Title("Password").
		Password(true).
		Value(&password).
		Run()

	return password, err
}

func loginAndPrintLastBill(username, password string) error {
	// login to PGE
	// get the last bill amount
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto("https://m.pge.com/#login"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}
	// Find the username input and fill it
	loc := page.Locator("input[name='username']")

	if err = loc.Fill(username); err != nil {
		log.Fatalf("could not fill username: %v", err)
	}

	// Find the password input and fill it
	loc = page.Locator("input[name='password']")
	if err = loc.Fill(password); err != nil {
		log.Fatalf("could not fill password: %v", err)
	}
	// log.Default().Println("filled password")

	// Find the login button and click it
	loc = page.Locator("button#home_login_submit")
	if err = loc.Click(); err != nil {
		log.Fatalf("could not click login: %v", err)
	}
	// Wait for the page to load

	// fmt.Println("waiting for page to load")
	loc = page.Locator("#billingSummarySection")

	loc.First().WaitFor()

	// fmt.Println("waiting for billing summary section")
	var text string
	text, err = loc.InnerText()
	if err != nil {
		log.Fatalf("could not get text: %v", err)
	}

	// Find first dollar amount in text
	dollarRegex := regexp.MustCompile(`\$\d+\.\d+`)
	dollarAmount := dollarRegex.FindString(text)
	if dollarAmount == "" {
		log.Fatalf("could not find dollar amount in text: %v", text)
	}
	fmt.Println(dollarAmount)

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
	return err
}

func PrettyJson(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
