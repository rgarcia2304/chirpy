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
	mux.Handle("/", http.FileServer(http.Dir(filePathRoot)))
	mux.Handle("/assets", http.FileServer(http.Dir("./assets/logo.png")))

	s := &http.Server{
		Addr: port,
		Handler: mux, 
	}

	fmt.Println("Serving files from %s on port %s\n,", filePathRoot, port)
	log.Fatal(s.ListenAndServe())
}
