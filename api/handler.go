package api

import (
	"encoding/json"
	"log"
	"music-lib/db"
	"music-lib/pkg/client"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	Repo   *db.Repository
	Client *client.SwaggerClient
}

// GetSongs godoc
// @Summary Get songs
// @Description Fetches a list of songs filtered by group and song name, with pagination support
// @Tags songs
// @Accept json
// @Produce json
// @Param group query string false "Filter by group name"
// @Param song query string false "Filter by song name"
// @Param limit query int false "Maximum number of songs to return (default: 10)"
// @Param offset query int false "Number of songs to skip (default: 0)"
// @Success 200 {array} db.Song "List of songs"
// @Failure 500 {string} string "Failed to fetch songs"
// @Router /songs [get]
func (h *Handler) GetSongs(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] Handling GetSongs request...")

	groupName := r.URL.Query().Get("group")
	songName := r.URL.Query().Get("song")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		log.Printf("[WARN] Invalid limit '%s', defaulting to 10", limitStr)
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		log.Printf("[WARN] Invalid offset '%s', defaulting to 0", offsetStr)
		offset = 0
	}

	if groupName != "" && len(groupName) < 3 {
		http.Error(w, "Group name is too short", http.StatusBadRequest)
		return
	}
	if songName != "" && len(songName) < 3 {
		http.Error(w, "Song name is too short", http.StatusBadRequest)
		return
	}

	songs, err := h.Repo.GetSongs(groupName, songName, limit, offset)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch songs: %v", err)
		http.Error(w, "Failed to fetch songs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)
}

// GetSongText godoc
// @Summary Get song text
// @Description Fetches the text of a song by its ID, with optional limit and offset
// @Tags songs
// @Accept json
// @Produce plain
// @Param id path int true "Song ID"
// @Success 200 {string} string "Song text"
// @Failure 400 {string} string "Invalid song ID"
// @Failure 500 {string} string "Failed to fetch song text"
// @Router /songs/{id}/text [get]

func (h *Handler) GetSongText(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] Handling GetSongText request...")
	vars := mux.Vars(r)

	songIDStr := vars["id"]
	songID, err := strconv.Atoi(songIDStr)
	if err != nil || songID <= 0 {
		log.Printf("[ERROR] Invalid song ID: %s", songIDStr)
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}

	text, err := h.Repo.GetSongText(songID)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch song text: %v", err)
		http.Error(w, "Failed to fetch song text", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(text))
}

// DeleteSong godoc
// @Summary Delete a song
// @Description Deletes a song by its ID
// @Tags songs
// @Accept json
// @Produce plain
// @Param id path int true "Song ID"
// @Success 200 {string} string "Song deleted successfully"
// @Failure 400 {string} string "Invalid song ID"
// @Failure 500 {string} string "Failed to delete song"
// @Router /songs/{id} [delete]
func (h *Handler) DeleteSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songIDStr := vars["id"]
	songID, err := strconv.Atoi(songIDStr)
	if err != nil || songID <= 0 {
		log.Printf("ERROR: Invalid song ID: %s", songIDStr)
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.DeleteSong(songID); err != nil {
		log.Printf("ERROR: Failed to delete song with ID %d: %v", songID, err)
		http.Error(w, "Failed to delete song", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Song deleted successfully"))
}

// UpdateSong godoc
// @Summary Update a song
// @Description Updates details of a song by its ID
// @Tags songs
// @Accept json
// @Produce plain
// @Param id path int true "Song ID"
// @Param song body db.Song true "Updated song details"
// @Success 200 {string} string "Song updated successfully"
// @Failure 400 {string} string "Invalid song ID or request body"
// @Failure 500 {string} string "Failed to update song"
// @Router /songs/{id} [put]
func (h *Handler) UpdateSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songIDStr := vars["id"]
	songID, err := strconv.Atoi(songIDStr)
	if err != nil || songID <= 0 {
		log.Printf("ERROR: Invalid song ID: %s", songIDStr)
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}

	var song db.Song
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		log.Printf("ERROR: Failed to decode request body for song ID %d: %v", songID, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Repo.UpdateSong(songID, song.GroupName, song.SongName, song.ReleaseDate, song.Text, song.Link); err != nil {
		log.Printf("ERROR: Failed to update song with ID %d: %v", songID, err)
		http.Error(w, "Failed to update song", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Song updated successfully"))
}

// AddSong godoc
// @Summary Add a new song
// @Description Adds a new song with details fetched from external API
// @Tags songs
// @Accept json
// @Produce plain
// @Param song body AddSongRequest true "Song details"
// @Success 201 {string} string "Song added successfully"
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Failed to fetch song details or save to DB"
// @Router /songs [post]
func (h *Handler) AddSong(w http.ResponseWriter, r *http.Request) {
	var input AddSongRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("ERROR: Failed to decode AddSongRequest: %v", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if len(input.Group) < 3 || len(input.Song) < 3 {
		http.Error(w, "Group name and song name must be at least 3 characters long", http.StatusBadRequest)
		return
	}

	songDetails, err := h.Client.FetchSongDetails(input.Group, input.Song)
	if err != nil {
		log.Printf("ERROR: Failed to fetch song details for group: %s, song: %s: %v", input.Group, input.Song, err)
		http.Error(w, "Failed to fetch song details", http.StatusInternalServerError)
		return
	}

	song := db.Song{
		GroupName:   input.Group,
		SongName:    input.Song,
		ReleaseDate: songDetails.ReleaseDate,
		Text:        songDetails.Text,
		Link:        songDetails.Link,
	}

	if err := h.Repo.CreateSong(song); err != nil {
		log.Printf("ERROR: Failed to save song to DB: %+v, Error: %v", song, err)
		http.Error(w, "Failed to save song to DB", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
