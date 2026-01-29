package main

import(
	"net/http"
	"github.com/google/uuid"
	"time"
	"encoding/json"
	"log"
	"github.com/rgarcia2304/chirpy/internal/database"
	"github.com/rgarcia2304/chirpy/internal/auth"
)

func (cfg *apiConfig) userHandler(w http.ResponseWriter, r *http.Request){
	type parameters struct{
		Password string `json:"password"`
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

	//check that all fields contain relevant info
	if params.Email == "" || params.Password == ""{
		cfg.respondWithError(w, 400, "Please provide both email and password.")
		return
	}

	//hash the password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil{
		log.Printf("Error hashing password %s", err)
		w.WriteHeader(500)
		return 
	}

	//now incorporate the methods to call the user
	createUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		HashedPassword: hashedPassword,
		Email: params.Email,
	})

	if err != nil{
		log.Printf("Error creating the user because %s", err)
		w.WriteHeader(500)
		return
	}
	
	
	resp := okResponse{ID: createUser.ID, CreatedAt: createUser.CreatedAt, UpdatedAt: createUser.UpdatedAt, Email: createUser.Email}
	log.Printf("This is the response %v", resp)
	cfg.respondWithJSON(w, 201, resp)

}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request){
	type parameters struct{
		Password string `json:"password"`
		Email string `json:"email"`
	}

	type response struct{
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`	
	}

	//unmarshall the parameter data
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil{
		log.Printf("Error marshalling data %s", err)
		w.WriteHeader(500)
		return 
	}

	//get the user 

	usr, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	//now check the password hashes
	valid, err := auth.CheckPasswordHash(params.Password, usr.HashedPassword)
	if err != nil{
		log.Printf("Error with comparing passwords %s", err)
		w.WriteHeader(500)
		return 
	}

	if valid{
		resp :=  response{ID: usr.ID, CreatedAt: usr.CreatedAt, UpdatedAt: usr.UpdatedAt, Email: usr.Email} 
		cfg.respondWithJSON(w, 200, resp)
	}else{
		cfg.respondWithError(w, 401, "Incorrect email or password")
	}

}
func (cfg *apiConfig) getChirpByIDHandler( w http.ResponseWriter, r *http.Request){

	type ChirpResp struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	parsedUUID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Fatalf("failed to parse UUID string: %v", err)
	}	
	log.Printf("The path value is ", r.PathValue("chirpID"))
	chirp, err := cfg.db.GetChirpByID(r.Context(), parsedUUID)
	if err != nil{
		cfg.respondWithError(w, 404, "Resource not found")
		return	
	}

	result := ChirpResp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}

	cfg.respondWithJSON(w, 200, result)
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

	type ChirpResp struct{
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
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
		resp := ChirpResp{ID: createdChirp.ID, CreatedAt: createdChirp.CreatedAt, UpdatedAt: createdChirp.UpdatedAt, Body: createdChirp.Body, UserID: createdChirp.ID}
		cfg.respondWithJSON(w, 201, resp)
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
		resp := ChirpResp{ID: createdChirp.ID, CreatedAt: createdChirp.CreatedAt, UpdatedAt: createdChirp.UpdatedAt, Body: createdChirp.Body, UserID: createdChirp.ID}	
		cfg.respondWithJSON(w, 201, resp)
	}	
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request){
	
	
	type ChirpResp struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	
	//get all the chirps resposne from the database
	chirpsLst, err := cfg.db.GetChirps(r.Context())
	if err != nil{
		log.Printf("Error creating the user because %s", err)
		w.WriteHeader(500)
		return
	}

	responses := make([]ChirpResp, len(chirpsLst))
	for i, c := range chirpsLst{
		responses[i] = ChirpResp{
			ID:        c.ID,
        		CreatedAt: c.CreatedAt,
        		UpdatedAt: c.UpdatedAt,
        		Body:      c.Body,
        		UserID:    c.UserID,
		}
	}
	cfg.respondWithJSON(w,200, responses)

}


