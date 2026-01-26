package main 
import(
	"net/http"
	"time"
	"log"
)

func main() {
	ServeMux := http.NewServeMux()

	s := &http.Server{
		Addr: ":8080",
		Handler: ServeMux, 
		ReadTimeout:  10 * time.Second, 
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20, 
	}

	log.Fatal(s.ListenAndServe())
}
