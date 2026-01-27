package main 

import(
	"net/http"
	"log"
	"fmt"
	"sync/atomic"
	"strconv"
)
type apiConfig struct{
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) requestsHandler(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		message := "Hits: " + strconv.FormatInt(int64(cfg.fileserverHits.Load()), 10)
		w.Write([]byte(message))
}

func (cfg *apiConfig) resetHandler( w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	message := "Reset Completed"
	w.Write([]byte(message))
}

func main() {
	const filePathRoot = "."
	const port = ":8080"
	
	//initialize the fileserverHits
	apiCfg := apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	//mux.Handle("/assets", http.FileServer(http.Dir("/assets/")))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ok"))
	})

	mux.HandleFunc("GET /metrics", apiCfg.requestsHandler)

	mux.HandleFunc("POST /reset", apiCfg.resetHandler)

	s := &http.Server{
		Addr: port,
		Handler: mux, 
	}

	fmt.Println("Serving files from %s on port %s\n,", filePathRoot, port)
	log.Fatal(s.ListenAndServe())
}
