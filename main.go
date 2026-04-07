package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handler_metric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>\n", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handler_reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Write([]byte(fmt.Sprintf("Hits: %d", 0)))
}

type validate_chirp struct {
	Body string `json:"body"`
}

type error_response struct {
	Error string `json:"error"`
}

type valid_response struct {
	Valid       bool   `json:"valid"`
	CleanedBody string `json:"cleaned_body"`
}

func cleanInput(s string, words []string) string {
	if len(words) == 0 {
		return s
	}
	pattern := `(?i)\b(` + strings.Join(words, "|") + `)\b`
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(s, "****")
	/*
		split_string := strings.Split(s, " ")
		for i, word := range split_string {
			for _, bad_word := range words {
				if strings.EqualFold(word, bad_word) {
					split_string[i] = "****"
				}
			}
		}
		return strings.Join(split_string, " ")
	*/
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, error_response{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func validate_chirp_handler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := validate_chirp{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Error decoding parameters")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	banned_words := []string{"kerfuffle", "sharbert", "fornax"}
	respondWithJSON(w, 200, valid_response{CleanedBody: cleanInput(params.Body, banned_words)})
}

func main() {
	const filepathRoot = "."
	const port = ":8080"
	var apiCfg apiConfig
	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	serverMux.HandleFunc("GET /api/healthz",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(200)
			w.Write([]byte("OK"))
		})
	serverMux.HandleFunc("GET /admin/metrics", apiCfg.handler_metric)
	serverMux.HandleFunc("POST /admin/reset", apiCfg.handler_reset)
	serverMux.HandleFunc("POST /api/validate_chirp", validate_chirp_handler)
	server := http.Server{
		Addr:    port,
		Handler: serverMux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
