package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Player struct {
	ID          int    `json:"id"`
	FullName    string `json:"full_name"`
	ShirtNumber int    `json:"shirt_number"`
	MarketValue string `json:"market_value"`
	BirthDate   string `json:"birth_date"`
	Position    string `json:"position"`
}

type Message struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

var players []Player

func main() {
	loadPlayers()

	http.HandleFunc("/api/ping", pingHandler)
	http.HandleFunc("/api/players/", playerByIDHandler)
	http.HandleFunc("/api/players", playersHandler)

	log.Println("Players API running on :24355")
	log.Fatal(http.ListenAndServe(":24355", nil))
}

func loadPlayers() {
	file, err := os.ReadFile("./data/players.json")
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	err = json.Unmarshal(file, &players)
	if err != nil {
		log.Fatal("Error parsing JSON:", err)
	}
}

func savePlayers() {
	data, err := json.MarshalIndent(players, "", "  ")
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	err = os.WriteFile("./data/players.json", data, 0644)
	if err != nil {
		log.Println("Error writing file:", err)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	response := Message{
		Message: "pong",
	}

	writeJSON(w, http.StatusOK, response)
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetPlayers(w, r)
	case http.MethodPost:
		handleCreatePlayer(w, r)
	default:
		writeErrorJSON(w, http.StatusMethodNotAllowed, "Method not allowed", "Allowed methods: GET, POST")
	}
}

func playerByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromPath(r.URL.Path, "/api/players/")
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid path parameter", err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetPlayerByID(w, id)
	case http.MethodPut:
		handleReplacePlayer(w, r, id)
	case http.MethodPatch:
		handlePatchPlayer(w, r, id)
	case http.MethodDelete:
		handleDeletePlayer(w, id)
	default:
		writeErrorJSON(w, http.StatusMethodNotAllowed, "Method not allowed", "Allowed methods: GET, PUT, PATCH, DELETE")
	}
}

func handleGetPlayers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	nameFilter := strings.ToLower(strings.TrimSpace(query.Get("name")))
	positionFilter := strings.ToLower(strings.TrimSpace(query.Get("position")))
	marketValueFilter := strings.ToLower(strings.TrimSpace(query.Get("market_value")))
	shirtNumberParam := strings.TrimSpace(query.Get("shirt_number"))

	var filtered []Player

	for _, player := range players {
		match := true

		if nameFilter != "" && !strings.Contains(strings.ToLower(player.FullName), nameFilter) {
			match = false
		}

		if positionFilter != "" && strings.ToLower(player.Position) != positionFilter {
			match = false
		}

		if marketValueFilter != "" && !strings.Contains(strings.ToLower(player.MarketValue), marketValueFilter) {
			match = false
		}

		if shirtNumberParam != "" {
			shirtNumber, err := strconv.Atoi(shirtNumberParam)
			if err != nil {
				writeErrorJSON(w, http.StatusBadRequest, "Invalid query parameter", "shirt_number must be an integer")
				return
			}
			if player.ShirtNumber != shirtNumber {
				match = false
			}
		}

		if match {
			filtered = append(filtered, player)
		}
	}

	writeJSON(w, http.StatusOK, filtered)
}

func handleGetPlayerByID(w http.ResponseWriter, id int) {
	index := findPlayerIndexByID(id)
	if index == -1 {
		writeErrorJSON(w, http.StatusNotFound, "Player not found", "No player exists with the provided id")
		return
	}

	writeJSON(w, http.StatusOK, players[index])
}

func handleCreatePlayer(w http.ResponseWriter, r *http.Request) {
	var newPlayer Player

	err := json.NewDecoder(r.Body).Decode(&newPlayer)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid JSON body", "Request body must be valid JSON")
		return
	}

	if validationError := validatePlayer(newPlayer, false); validationError != "" {
		writeErrorJSON(w, http.StatusBadRequest, "Validation error", validationError)
		return
	}

	newPlayer.ID = generateNextID()
	players = append(players, newPlayer)
	savePlayers()

	writeJSON(w, http.StatusCreated, newPlayer)
}

func handleReplacePlayer(w http.ResponseWriter, r *http.Request, id int) {
	index := findPlayerIndexByID(id)
	if index == -1 {
		writeErrorJSON(w, http.StatusNotFound, "Player not found", "No player exists with the provided id")
		return
	}

	var updatedPlayer Player
	err := json.NewDecoder(r.Body).Decode(&updatedPlayer)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid JSON body", "Request body must be valid JSON")
		return
	}

	if validationError := validatePlayer(updatedPlayer, false); validationError != "" {
		writeErrorJSON(w, http.StatusBadRequest, "Validation error", validationError)
		return
	}

	updatedPlayer.ID = id
	players[index] = updatedPlayer
	savePlayers()

	writeJSON(w, http.StatusOK, updatedPlayer)
}

func handlePatchPlayer(w http.ResponseWriter, r *http.Request, id int) {
	index := findPlayerIndexByID(id)
	if index == -1 {
		writeErrorJSON(w, http.StatusNotFound, "Player not found", "No player exists with the provided id")
		return
	}

	var updates map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid JSON body", "Request body must be valid JSON")
		return
	}

	player := players[index]

	for key, value := range updates {
		switch key {
		case "full_name":
			stringValue, ok := value.(string)
			if !ok || strings.TrimSpace(stringValue) == "" {
				writeErrorJSON(w, http.StatusBadRequest, "Validation error", "full_name must be a non-empty string")
				return
			}
			player.FullName = strings.TrimSpace(stringValue)

		case "shirt_number":
			floatValue, ok := value.(float64)
			if !ok {
				writeErrorJSON(w, http.StatusBadRequest, "Validation error", "shirt_number must be a number")
				return
			}
			player.ShirtNumber = int(floatValue)

		case "market_value":
			stringValue, ok := value.(string)
			if !ok || strings.TrimSpace(stringValue) == "" {
				writeErrorJSON(w, http.StatusBadRequest, "Validation error", "market_value must be a non-empty string")
				return
			}
			player.MarketValue = strings.TrimSpace(stringValue)

		case "birth_date":
			stringValue, ok := value.(string)
			if !ok || strings.TrimSpace(stringValue) == "" {
				writeErrorJSON(w, http.StatusBadRequest, "Validation error", "birth_date must be a non-empty string")
				return
			}
			player.BirthDate = strings.TrimSpace(stringValue)

		case "position":
			stringValue, ok := value.(string)
			if !ok || strings.TrimSpace(stringValue) == "" {
				writeErrorJSON(w, http.StatusBadRequest, "Validation error", "position must be a non-empty string")
				return
			}
			player.Position = strings.ToLower(strings.TrimSpace(stringValue))

		case "id":
			writeErrorJSON(w, http.StatusBadRequest, "Validation error", "id cannot be modified")
			return

		default:
			writeErrorJSON(w, http.StatusBadRequest, "Validation error", "Unknown field: "+key)
			return
		}
	}

	if validationError := validatePlayer(player, true); validationError != "" {
		writeErrorJSON(w, http.StatusBadRequest, "Validation error", validationError)
		return
	}

	players[index] = player
	savePlayers()

	writeJSON(w, http.StatusOK, player)
}

func handleDeletePlayer(w http.ResponseWriter, id int) {
	index := findPlayerIndexByID(id)
	if index == -1 {
		writeErrorJSON(w, http.StatusNotFound, "Player not found", "No player exists with the provided id")
		return
	}

	deletedPlayer := players[index]
	players = append(players[:index], players[index+1:]...)
	savePlayers()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Player deleted successfully",
		"player":  deletedPlayer,
	})
}

func validatePlayer(player Player, allowPartial bool) string {
	validPositions := map[string]bool{
		"portero":       true,
		"defensa":       true,
		"mediocampista": true,
		"delantero":     true,
	}

	if strings.TrimSpace(player.FullName) == "" {
		return "full_name is required"
	}

	if player.ShirtNumber <= 0 {
		return "shirt_number must be greater than 0"
	}

	if strings.TrimSpace(player.MarketValue) == "" {
		return "market_value is required"
	}

	if strings.TrimSpace(player.BirthDate) == "" {
		return "birth_date is required"
	}

	if strings.TrimSpace(player.Position) == "" {
		return "position is required"
	}

	if !validPositions[strings.ToLower(player.Position)] {
		return "position must be one of: portero, defensa, mediocampista, delantero"
	}

	return ""
}

func generateNextID() int {
	maxID := 0

	for _, player := range players {
		if player.ID > maxID {
			maxID = player.ID
		}
	}

	return maxID + 1
}

func findPlayerIndexByID(id int) int {
	for index, player := range players {
		if player.ID == id {
			return index
		}
	}
	return -1
}

func extractIDFromPath(path string, prefix string) (int, error) {
	idPart := strings.TrimPrefix(path, prefix)
	idPart = strings.TrimSpace(idPart)

	if idPart == "" {
		return 0, http.ErrMissingFile
	}

	if strings.Contains(idPart, "/") {
		return 0, strconv.ErrSyntax
	}

	id, err := strconv.Atoi(idPart)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func writeErrorJSON(w http.ResponseWriter, status int, message string, details string) {
	writeJSON(w, status, ErrorResponse{
		Error:   message,
		Details: details,
	})
}