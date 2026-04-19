package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mike-the-math-man/Chirpy.git/internal/auth"
	"github.com/mike-the-math-man/Chirpy.git/internal/database"
)

type apiConfig struct {
	fileserverHits  atomic.Int32
	databaseQueries *database.Queries
	env_platform    string
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
	if cfg.env_platform != "dev" {
		w.WriteHeader(403)
		return
	}
	cfg.fileserverHits.Store(0)
	w.Write([]byte(fmt.Sprintf("Hits: %d\n", 0)))
	err := cfg.databaseQueries.DeleteUsers(r.Context())
	if err != nil {
		fmt.Printf("error deleting users: %v\n", err)
	}
}

type validate_chirp struct {
	Body string `json:"body"`
}

type full_chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

type error_response struct {
	Error string `json:"error"`
}

type valid_response struct {
	Valid       bool   `json:"valid"`
	CleanedBody string `json:"cleaned_body"`
}

type user_email_password struct {
	Email          string `json:"email"`
	HashedPassword string `json:"password"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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

func (cfg *apiConfig) users_handler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := user_email_password{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Error decoding parameters")
		return
	}
	if len(params.HashedPassword) < 1 {
		fmt.Println("please provide password")
		respondWithError(w, 400, "please provide password")
	}
	params.HashedPassword, err = auth.HashPassword(params.HashedPassword)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		respondWithError(w, 500, "Error hashing password")
		return
	}
	user, err := cfg.databaseQueries.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: params.HashedPassword})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondWithError(w, 500, "Error creating user")
		return
	}
	user_struct := User{}
	user_struct.CreatedAt = user.CreatedAt
	user_struct.UpdatedAt = user.UpdatedAt
	user_struct.Email = user.Email
	user_struct.ID = user.ID
	respondWithJSON(w, 201, user_struct)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := user_email_password{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Error decoding parameters")
		return
	}
	user, err := cfg.databaseQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		fmt.Printf("Error getting hashed password %v\n", err)
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	valid, err := auth.CheckPasswordHash(params.HashedPassword, user.HashedPassword)
	if err != nil {
		fmt.Printf("Error validating hashed password %v\n", err)
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	user_struct := User{}
	user_struct.CreatedAt = user.CreatedAt
	user_struct.UpdatedAt = user.UpdatedAt
	user_struct.Email = user.Email
	user_struct.ID = user.ID
	if valid {
		respondWithJSON(w, 200, user_struct)
	} else {
		respondWithError(w, 401, "Incorrect email or password")
	}

}

/*
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
*/
func (cfg *apiConfig) chirps_handler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := full_chirp{}
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
	params.Body = cleanInput(params.Body, banned_words)
	chirp, err := cfg.databaseQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: params.Body, UserID: params.UserId})
	chirp_data := full_chirp{}
	chirp_data.Id = chirp.ID
	chirp_data.Body = chirp.Body
	chirp_data.CreatedAt = chirp.CreatedAt
	chirp_data.UpdatedAt = chirp.UpdatedAt
	chirp_data.UserId = chirp.UserID
	respondWithJSON(w, 201, chirp_data)
}

func (cfg *apiConfig) chirps_get_handler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.databaseQueries.GetChirps(r.Context())
	if err != nil {
		fmt.Printf("error getting chirps %v\n", err)
		respondWithError(w, 500, "Error getting chirps")
		return
	}
	chirp_data_list := []full_chirp{}
	for _, chirp := range chirps {
		chirp_data := full_chirp{}
		chirp_data.Id = chirp.ID
		chirp_data.Body = chirp.Body
		chirp_data.CreatedAt = chirp.CreatedAt
		chirp_data.UpdatedAt = chirp.UpdatedAt
		chirp_data.UserId = chirp.UserID
		chirp_data_list = append(chirp_data_list, chirp_data)
	}

	respondWithJSON(w, 200, chirp_data_list)
}

func (cfg *apiConfig) chirps_get_individual_handler(w http.ResponseWriter, r *http.Request) {
	chirp_user_id_string := r.PathValue("chirpID")
	chirp_user_id_uuid, err := uuid.Parse(chirp_user_id_string)
	if err != nil {
		fmt.Printf("error parsing user_id %v\n", err)
		respondWithError(w, 500, "")
		return
	}
	chirp, err := cfg.databaseQueries.GetChirp(r.Context(), chirp_user_id_uuid)
	if err != nil {
		fmt.Printf("error getting chirp %v\n", err)
		respondWithError(w, 404, "")
		return
	}
	chirp_data := full_chirp{}
	chirp_data.Id = chirp.ID
	chirp_data.Body = chirp.Body
	chirp_data.CreatedAt = chirp.CreatedAt
	chirp_data.UpdatedAt = chirp.UpdatedAt
	chirp_data.UserId = chirp.UserID

	respondWithJSON(w, 200, chirp_data)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	if err != nil {
		return
	}
	const filepathRoot = "."
	const port = ":8080"
	var apiCfg apiConfig
	platform := os.Getenv("PLATFORM")
	apiCfg.env_platform = platform
	apiCfg.databaseQueries = dbQueries
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
	//serverMux.HandleFunc("POST /api/validate_chirp", validate_chirp_handler)
	serverMux.HandleFunc("POST /api/users", apiCfg.users_handler)
	serverMux.HandleFunc("POST /api/chirps", apiCfg.chirps_handler)
	serverMux.HandleFunc("GET /api/chirps", apiCfg.chirps_get_handler)
	serverMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.chirps_get_individual_handler)
	serverMux.HandleFunc("POST /api/login", apiCfg.login)
	server := http.Server{
		Addr:    port,
		Handler: serverMux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
