package main 
import(
	"net/http"
	"log"
	"fmt"
)

func main() {
	const filePathRoot = "."
	const port = ":8080"

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot))))
	//mux.Handle("/assets", http.FileServer(http.Dir("/assets/")))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ok"))
	})

	s := &http.Server{
		Addr: port,
		Handler: mux, 
	}

	fmt.Println("Serving files from %s on port %s\n,", filePathRoot, port)
	log.Fatal(s.ListenAndServe())
}
