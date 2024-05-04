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
	"log/slog"
	"os"
	"regexp"
	"strconv"

	"github.com/anvari1313/splitwise.go"
	"github.com/charmbracelet/charm/kv"
	"github.com/charmbracelet/huh"
	"github.com/playwright-community/playwright-go"
	"github.com/urfave/cli/v2"
	"github.com/whoaa512/cjs-tools/pkg/collections"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
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

					dollarAmt, err := getLastPgeDollarAmount(username, password)
					if err != nil {
						return err
					}

					logger.Info("Last bill amount", "dollarAmt", dollarAmt)

					return err
				},
			},
			{
				Name:  "add-expense",
				Usage: "Add an expense to Splitwise",
				Action: func(c *cli.Context) error {
					var err error
					db, err := kv.OpenWithDefaults("charm.sh.kv.user.default")
					if err != nil {
						return err
					}

					v, _ := db.Get([]byte("splitwise_api_key"))

					apiKey := string(v)
					if apiKey == "" {
						apiKey = os.Getenv("SPLITWISE_API_KEY")
					}
					if apiKey == "" {
						err = huh.NewInput().Title("Splitwise API Key").Value(&apiKey).Run()
					}
					if err != nil {
						return err
					}
					if apiKey != "" {
						_ = db.Set([]byte("splitwise_api_key"), []byte(apiKey))
					} else {
						return fmt.Errorf("Splitwise API Key is required")
					}
					client := splitwise.NewClient(
						splitwise.NewAPIKeyAuth(apiKey),
					)

					rawGroups, err := client.Groups(c.Context)
					if err != nil {
						return err
					}

					groupOptions := collections.Map(func(g splitwise.Group) huh.Option[int] {
						return huh.Option[int]{Key: g.Name, Value: int(g.ID)}
					}, rawGroups)

					var chosenGroupId int
					v, _ = db.Get([]byte("last_used_splitwise_group_id"))
					if v != nil {
						prevGroupId, err := strconv.Atoi(string(v))
						if err == nil {
							chosenGroupId = prevGroupId
						}
					}

					if chosenGroupId == 0 {
						err = huh.NewSelect[int]().
							Title("Select Group").
							Options(groupOptions...).
							Validate(func(i int) error {
								if i == 0 {
									return fmt.Errorf("Group selection is required")
								}
								return nil
							}).
							Value(&chosenGroupId).
							Run()

						if err != nil {
							return err
						}
					}

					group, ok := collections.Find(func(g splitwise.Group) bool {
						return int(g.ID) == chosenGroupId
					}, rawGroups)
					if !ok {
						return fmt.Errorf("Group not found")
					}

					exps, err := client.CreateExpenseSplitEqually(
						c.Context,
						splitwise.ExpenseSplitEqually{
							Expense: splitwise.Expense{
								Cost:         "420.69",
								GroupId:      uint32(group.ID),
								CurrencyCode: "USD",
								Date:         "2024-04-15",
								Description:  "CJ testing",
							},
							SplitEqually: true,
						},
					)
					if err != nil {
						return fmt.Errorf("Error creating expense: %w", err)
					}

					fmt.Printf("%v", PrettyJson(exps))
					// https://github.com/anvari1313/splitwise.go/blob/main/groups.go

					return err
				},
			},
		},
		// Command to use the Splitwise API to add an expense
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

func getLastPgeDollarAmount(username, password string) (string, error) {
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

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
	return dollarAmount, err
}

func PrettyJson(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		b, _ = json.MarshalIndent(
			map[string]string{"error": "error marshalling json: " + err.Error()},
			"",
			"  ",
		)
	}
	return string(b)
}

func addSplitwiseExpense(dollarAmt string) error {
	return nil
}
