package main

import(
	"net/http"
	"github.com/google/uuid"
	"time"
	"encoding/json"
	"log"
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
	
	
	resp := okResponse{ID: createUser.ID.UUID, CreatedAt: createUser.CreatedAt, UpdatedAt: createUser.UpdatedAt, Email: createUser.Email}
	log.Printf("This is the response %v", resp)
	cfg.respondWithJSON(w, 200, resp)

}
