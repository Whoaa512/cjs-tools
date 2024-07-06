package main

/* Quicklinks server
- Provides a simple web interface to show the list of links
- Add a new link
- Delete a link
- Update a link
- Search for a link

Frontend is all HTMX and TailwindCSS
Backend is Go with a SQLite database
*/

import (
	_ "github.com/mattn/go-sqlite3"
)

// Link struct
type Link struct {
	Id   int
	Name string
	Url  string
}

// LinksDatabase struct
type LinksDatabase struct {
	Links []Link
}

// main function
func main() {
	// Create a new database
	db := LinksDatabase{}
	db.Links = []Link{}
}

// Add a link to the database
func (db *LinksDatabase) AddLink(name string, url string) {
	// Create a new link
	link := Link{Name: name, Url: url}
	// Add the link to the database
	db.Links = append(db.Links, link)
}

// Remove a link from the database
func (db *LinksDatabase) RemoveLink(name string) {
	// Find the link in the database
	for i, link := range db.Links {
		if link.Name == name {
			// Remove the link from the database
			db.Links = append(db.Links[:i], db.Links[i+1:]...)
			return
		}
	}
}

// Update a link in the database
func (db *LinksDatabase) UpdateLink(name string, url string) {
	// Find the link in the database
	for i, link := range db.Links {
		if link.Name == name {
			// Update the link in the database
			db.Links[i].Url = url
			return
		}
	}
}

// Search for a link in the database
func (db *LinksDatabase) SearchLink(name string) *Link {
	// Find the link in the database
	for _, link := range db.Links {
		if link.Name == name {
			return &link
		}
	}
	return nil
}

// List all links in the database
func (db *LinksDatabase) ListLinks() []Link {
	return db.Links
}

// Save the links to the database
func (db *LinksDatabase) SaveLinks() {
	// Save the links to the database
}
