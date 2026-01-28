package main 

import(
	"net/http"
	"log"
	"fmt"
	"sync/atomic"
	"encoding/json"
)
type apiConfig struct{
	fileserverHits atomic.Int32
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
	data, err := json.Marshal(payload)
	if err != nil{
		log.Printf("Error marshalling JSON %s", err)
		w.WriteHeader(500)
		return
	}
	
	w.WriteHeader(code)
	w.Write(data)
}

func(cfg *apiConfig) validateChirpHandler(w http.ResponseWriter, r *http.Request){
	type parameters struct{
		Body string `json:"body"`
	}
	
	type okResp struct{
		Body string `json:"body"`
	}

	type cleanedResp struct{
		CleanedBody string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil{
		log.Printf("Error marshalling data %s", err)
		w.WriteHeader(500)
		return
	}

	if len(params.Body) > 140{
		cfg.respondWithError(w , 400, "Chirp is too long")
		return
	}
	
	msg, profane := profaneCheck(params.Body)
	if profane{
		resp := cleanedResp{CleanedBody: msg}
		cfg.respondWithJSON(w, 200, resp)
	}else{	
		resp := okResp{Body: params.Body}
		cfg.respondWithJSON(w, 200, resp)
	}
}


func main() {
	const filePathRoot = "."
	const port = ":8080"
	
	//initialize the fileserverHits
	apiCfg := apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	//mux.Handle("/assets", http.FileServer(http.Dir("/assets/")))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ok"))
	})

	mux.HandleFunc("GET /admin/metrics", apiCfg.requestsHandler)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	mux.HandleFunc("POST /api/validate_chirp", apiCfg.validateChirpHandler)

	s := &http.Server{
		Addr: port,
		Handler: mux, 
	}

	fmt.Println("Serving files from %s on port %s\n,", filePathRoot, port)
	log.Fatal(s.ListenAndServe())
}
