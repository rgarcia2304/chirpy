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

func main() {
	const filePathRoot = "."
	const port = ":8080"
	
	//initialize the fileserverHits
	apiCfg := apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	//mux.Handle("/assets", http.FileServer(http.Dir("/assets/")))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ok"))
	})

	mux.HandleFunc("/metrics", apiCfg.requestsHandler)

	s := &http.Server{
		Addr: port,
		Handler: mux, 
	}

	fmt.Println("Serving files from %s on port %s\n,", filePathRoot, port)
	log.Fatal(s.ListenAndServe())
}
