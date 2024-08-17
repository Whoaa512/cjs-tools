package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

//go:embed index.html
//go:embed results.html
var content embed.FS

type Video struct {
	ID          string
	Title       string
	PublishedAt time.Time
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/search", handleSearch)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(content, "index.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	startDate, err := time.Parse("2006-01-02", r.FormValue("startDate"))
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse("2006-01-02", r.FormValue("endDate"))
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	videos, err := getYouTubeVideos(query, startDate, endDate)
	if err != nil {
		http.Error(w, "Error fetching videos", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFS(content, "results.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, videos)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func getYouTubeVideos(query string, startDate, endDate time.Time) ([]Video, error) {
	// TODO: Replace with your actual YouTube API key
	apiKey := "YOUR_YOUTUBE_API_KEY"

	service, err := youtube.NewService(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creating YouTube client: %v", err)
	}

	call := service.Search.List([]string{"id", "snippet"}).
		Q(query).
		Type("video").
		PublishedAfter(startDate.Format(time.RFC3339)).
		PublishedBefore(endDate.Format(time.RFC3339)).
		MaxResults(10)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("error executing search: %v", err)
	}

	var videos []Video
	for _, item := range response.Items {
		publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		videos = append(videos, Video{
			ID:          item.Id.VideoId,
			Title:       item.Snippet.Title,
			PublishedAt: publishedAt,
		})
	}

	return videos, nil
}
