package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type Link struct {
	ShortKey    string `json:"key"`
	OriginalURL string `json:"url"`
	Clicks      int    `json:"clicks"`
}

var links map[string]Link
var mu sync.Mutex

func createLink(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Ошибка, я жду только POST", http.StatusMethodNotAllowed)
		return
	}
	type AnonLink struct {
		Url   string `json:"url"`
		Alias string `json:"alias"`
	}

	var anon AnonLink

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&anon)
	if err != nil {
		http.Error(w, "Неправильный JSON", http.StatusBadRequest)
		return
	}

	if anon.Url == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if anon.Alias == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	mu.Lock()
	_, ok := links[anon.Alias]

	if ok {
		mu.Unlock()
		http.Error(w, "Данное имя занято", http.StatusBadRequest)
		return
	}

	NewLink := Link{
		ShortKey:    anon.Alias,
		OriginalURL: anon.Url,
		Clicks:      0,
	}

	links[anon.Alias] = NewLink
	saveFiles("url.json", links)
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Ссылка создана: /%s", anon.Alias)

}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]

	mu.Lock()
	linkData, ok := links[key]

	if !ok {
		mu.Unlock()
		http.Error(w, "Введите ссылку", http.StatusNotFound)
		return
	}
	linkData.Clicks++
	links[key] = linkData
	saveFiles("url.json", links)
	mu.Unlock()

	http.Redirect(w, r, linkData.OriginalURL, http.StatusFound)
}

func main() {
	links = make(map[string]Link)

	links = loadFiles("url.json")
	if len(links) == 0 {
		links["gw"] = Link{
			ShortKey:    "gw",
			OriginalURL: "https://google.com",
			Clicks:      0,
		}
		saveFiles("url.json", links)
	}
	fmt.Println("Server listening on port 8080...")

	http.HandleFunc("/create", createLink)
	http.HandleFunc("/", redirectHandler)
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Error", err)
	}

	fmt.Println("Server listening on port 8080...")
}

func saveFiles(fileName string, data map[string]Link) {

	fileData, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		fmt.Println("Ошибка кодирования jSON файла", err)

		return
	}
	err = os.WriteFile(fileName, fileData, 0644)

	if err != nil {
		fmt.Println("Ошибка создания JSON файла", err)
	}
}

func loadFiles(fileName string) map[string]Link {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return map[string]Link{}
	}

	data, err := os.ReadFile(fileName)

	if err != nil {
		fmt.Println("Ошибка чтения JSON:", err)

		return map[string]Link{}
	}

	var loadedData map[string]Link

	err = json.Unmarshal(data, &loadedData)

	if err != nil {
		fmt.Println("Ошибка декодирования JSON")
		return map[string]Link{}
	}

	return loadedData
}
