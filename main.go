package main 

import _ "github.com/lib/pq"

import(
	"net/http"
	"log"
	"fmt"
	"sync/atomic"
	"encoding/json"
	"os"
	"github.com/rgarcia2304/chirpy/internal/database"
	"database/sql"
	"github.com/joho/godotenv"
)
type apiConfig struct{
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
	jwtSecret string
}

func (cfg *apiConfig) requestsHandler(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		message := fmt.Sprintf("<html><body><h1>Welcome Chirpy, Admin </h1><p>Chirpy has visited %d</p></body></html>", cfg.fileserverHits.Load())
		w.Write([]byte(message))
}

func (cfg *apiConfig) resetHandler( w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	message := "Reset Completed"
	w.Write([]byte(message))
}

func(cfg *apiConfig) respondWithError(w http.ResponseWriter, code int, msg string){
	type errResp struct{
		Error string `json: "error"`
	}

	respBody := errResp{Error: msg}
	data, err := json.Marshal(respBody)
	if err != nil{
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
	return
}

func(cfg *apiConfig) respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
	w.Header().Set("Content-Type", "application/json")	
	data, err := json.Marshal(payload)
	if err != nil{
		log.Printf("Error marshalling JSON %s", err)
		w.WriteHeader(500)
		return
	}
	
	
    	log.Printf("responding %d with: %s", code, string(data))
	w.WriteHeader(code)
	w.Write(data)
	w.Write([]byte("\n"))
}


func main() {
	//load the env file 
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	plat := os.Getenv("PLATFORM")
	jwtScrt := os.Getenv("JWT_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil{
		fmt.Println(err)
	}
	dbQueries := database.New(db)

	const filePathRoot = "."
	const port = ":8080"
	
	//initialize the fileserverHits
	apiCfg := apiConfig{db: dbQueries, platform: plat, jwtSecret: jwtScrt}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	//mux.Handle("/assets", http.FileServer(http.Dir("/assets/")))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ok"))
	})

	mux.HandleFunc("GET /admin/metrics", apiCfg.requestsHandler)

	mux.HandleFunc("POST /admin/reset", apiCfg.userDeleteHandler)

	mux.HandleFunc("POST /api/users", apiCfg.userHandler)

	mux.HandleFunc("POST /api/chirps", apiCfg.createChirpsHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpByIDHandler)
	mux.HandleFunc("POST /api/login", apiCfg.loginHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeHandler)
	mux.HandleFunc("PUT /api/users", apiCfg.updateUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirp)



	s := &http.Server{
		Addr: port,
		Handler: mux, 
	}

	fmt.Println("Serving files from %s on port %s\n,", filePathRoot, port)
	log.Fatal(s.ListenAndServe())
}
