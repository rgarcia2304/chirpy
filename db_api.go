package main

import(
	"net/http"
	"github.com/google/uuid"
	"time"
	"encoding/json"
	"log"
	"github.com/rgarcia2304/chirpy/internal/database"

)

func (cfg *apiConfig) userHandler(w http.ResponseWriter, r *http.Request){
	type parameters struct{
		Email string `json:"email"`		
	}
	
	type okResponse struct{
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
	}
	
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil{
		log.Printf("Error marshalling data %s", err)
		w.WriteHeader(500)
		return 
	}

	//now incorporate the methods to call the user
	createUser, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil{
		log.Printf("Error creating the user because %s", err)
		w.WriteHeader(500)
		return
	}
	
	
	resp := okResponse{ID: createUser.ID, CreatedAt: createUser.CreatedAt, UpdatedAt: createUser.UpdatedAt, Email: createUser.Email}
	log.Printf("This is the response %v", resp)
	cfg.respondWithJSON(w, 201, resp)

}

func (cfg *apiConfig) userDeleteHandler(w http.ResponseWriter, r *http.Request){
	//call the delete directly
	if cfg.platform != "dev"{
		w.WriteHeader(403)
		return 
	}
	err := cfg.db.DeleteUsers(r.Context())
	if err != nil{
		log.Printf("Error creating the user because %s", err)
		w.WriteHeader(500)
		return
	}

}

func (cfg *apiConfig) createChirpsHandler(w http.ResponseWriter, r *http.Request){
	type parameters struct{
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
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
		//create the chirp in the database
		createdChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body: msg, 
			UserID: params.UserID,
		})
		
		if err != nil{
			log.Printf("Error creating the chirp because %s", err)
			w.WriteHeader(500)
			return
		}

		cfg.respondWithJSON(w, 201, createdChirp)
	}else{	
		createdChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body: params.Body, 
			UserID: params.UserID,
		})

		if err != nil{
			log.Printf("Error creating the user because %s", err)
			w.WriteHeader(500)
			return
		}
		
		cfg.respondWithJSON(w, 201, createdChirp)
	}	
}




