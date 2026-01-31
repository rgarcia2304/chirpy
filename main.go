package main 

import _ "github.com/lib/pq"

import(
	"net/http"
	"log"
	"fmt"
	"sync/atomic"
	"os"
	"github.com/rgarcia2304/chirpy/internal/database"
	"github.com/rgarcia2304/chirpy/internal/handlers"
	"database/sql"
	"github.com/joho/godotenv"
)
type ApiConfig struct{
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
	jwtSecret string
	polkaKey string
}




func main() {
	//load the env file 
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	plat := os.Getenv("PLATFORM")
	polka := os.Getenv("POLKA_KEY")
	jwtScrt := os.Getenv("JWT_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil{
		fmt.Println(err)
	}
	dbQueries := database.New(db)

	const filePathRoot = "."
	const port = ":8080"
	
	//initialize the fileserverHits
	apiCfg := handlers.ApiConfig{
	DB:        dbQueries,
	Platform:  plat,
	JWTSecret: jwtScrt,
	PolkaKey:  polka,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.MiddlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	//mux.Handle("/assets", http.FileServer(http.Dir("/assets/")))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ok"))
	})

	mux.HandleFunc("GET /admin/metrics", apiCfg.RequestsHandler)

	mux.HandleFunc("POST /admin/reset", apiCfg.UserDeleteHandler)

	mux.HandleFunc("POST /api/users", apiCfg.UserHandler)

	mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirpsHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirpByIDHandler)
	mux.HandleFunc("POST /api/login", apiCfg.LoginHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.RefreshHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.RevokeHandler)
	mux.HandleFunc("PUT /api/users", apiCfg.UpdateUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.DeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.UpgradeUser)

	s := &http.Server{
		Addr: port,
		Handler: mux, 
	}

	fmt.Println("Serving files from %s on port %s\n,", filePathRoot, port)
	log.Fatal(s.ListenAndServe())
}
