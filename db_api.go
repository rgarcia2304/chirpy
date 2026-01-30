package main

import(
	"net/http"
	"github.com/google/uuid"
	"time"
	"encoding/json"
	"log"
	"github.com/rgarcia2304/chirpy/internal/database"
	"github.com/rgarcia2304/chirpy/internal/auth"
	"strings"
	"fmt"
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
		IsChirpyRed bool `json:"is_chirpy_red"`
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
	
	
	resp := okResponse{ID: createUser.ID, CreatedAt: createUser.CreatedAt, UpdatedAt: createUser.UpdatedAt, Email: createUser.Email, IsChirpyRed: createUser.IsChirpyRed.Bool}
	log.Printf("This is the response %v", resp)
	cfg.respondWithJSON(w, 201, resp)

}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request){
	authHeader := r.Header.Get("Authorization")
	if authHeader == ""{
		cfg.respondWithError(w, 401, "missing authorization header")
		return
	}

	const prefix = "Bearer "
	
	if !strings.HasPrefix(authHeader, prefix){
		cfg.respondWithError(w, 401, "invalid authorization header")
		return
	}
	refreshToken := strings.TrimPrefix(authHeader, prefix)

	_, err := cfg.db.MarkTokenRevoked(r.Context(), refreshToken)
	if err != nil{
		log.Printf("The error is %v", err)
		cfg.respondWithError(w, 401, "Error Revoking Token")
		return	
	}
	w.WriteHeader(204)
	return
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request){
	//get the refreshToken
	//Once the refresh token is received, look it up in the database
	//if the toke is found return the token associated with it, if not return 401

	type okResp struct{
		Token string `json:"token"`
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == ""{
		cfg.respondWithError(w, 401, "missing authorization header")
		return
	}

	const prefix = "Bearer "
	
	if !strings.HasPrefix(authHeader, prefix){
		cfg.respondWithError(w, 401, "invalid authorization header")
		return
	}
	refreshToken := strings.TrimPrefix(authHeader, prefix)
	foundToken, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)
	if err != nil{
		cfg.respondWithError(w, 401, "Token does not exist")
		return
	}

	if time.Now().After(foundToken.ExpiresAt) {
		cfg.respondWithError(w, 401, "Refresh Token is expired")
		return
	}

	//generate the token for the user 
	//check if there is a time field
	tokenStr, err := auth.MakeJWT(foundToken.UserID, cfg.jwtSecret, 1 * time.Hour)
	if err != nil{
		cfg.respondWithError(w, 500, "Error making JWT Token")
		return
	}
	
	resp := okResp{Token: tokenStr}
	cfg.respondWithJSON(w, 200, resp)
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
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed bool `json:"is_chirpy_red"`
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
	
	//generate the token for the user 
	//check if there is a time field
	tokenStr, err := auth.MakeJWT(usr.ID, cfg.jwtSecret, 1 * time.Hour)
	if err != nil{
		cfg.respondWithError(w, 500, "Error making JWT Token")
		return
	}
	if valid{
		//make the refresh token
		rfrshTkn, err := auth.MakeRefreshToken()
		if err != nil{
			cfg.respondWithError(w, 500, "Issue making token")
			return
		}
		
		//register the refresh token
		newRefresh, err := cfg.db.CreateRefresh(r.Context(), database.CreateRefreshParams{
			Token: rfrshTkn, 
			UserID: usr.ID,
		})

		if err != nil{
			log.Printf("The error is %v", err)
			cfg.respondWithError(w, 500, "Issue making token")
			return
		}

		resp :=  response{ID: usr.ID, CreatedAt: usr.CreatedAt, UpdatedAt: usr.UpdatedAt, Email: usr.Email, Token: tokenStr, RefreshToken: newRefresh.Token, IsChirpyRed: usr.IsChirpyRed.Bool} 
		cfg.respondWithJSON(w, 200, resp)
		return
	}else{
		cfg.respondWithError(w, 401, "Incorrect email or password")
		return
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
		IsChirpyRed bool `json:"is_chirpy_red"`
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
	
	//Have to check that the user has a bearerID that is valid 
	//get the bearer token from the header
	//then validate that bearer token 
	authHeader, err := auth.GetBearerToken(r.Header)
	if err != nil{
		cfg.respondWithError(w, 500, "Error getting Authorization Header")
		return
	}
	_, err = auth.ValidateJWT(authHeader, cfg.jwtSecret)
	if err != nil{
		cfg.respondWithError(w, 401, "Unauthorized")
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
		resp := ChirpResp{ID: createdChirp.ID, CreatedAt: createdChirp.CreatedAt, UpdatedAt: createdChirp.UpdatedAt, Body: createdChirp.Body, UserID: createdChirp.UserID}
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
		resp := ChirpResp{ID: createdChirp.ID, CreatedAt: createdChirp.CreatedAt, UpdatedAt: createdChirp.UpdatedAt, Body: createdChirp.Body, UserID: createdChirp.UserID}	
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
		IsChirpyRed bool `json:"is_chirpy_red"`
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

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request){
	//grab access token
	//then provide new email and password
	type parameters struct{
		Password string `json:"password"`
		Email string `json:"email"`
	}

	type okResponse struct{
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
		IsChirpyRed bool `json:"is_chirpy_red"`
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

	//Have to check that the user has a bearerID that is valid 
	//get the bearer token from the header
	//then validate that bearer token 
	authHeader := r.Header.Get("Authorization")
	if authHeader == ""{
		cfg.respondWithError(w, 401, "missing authorization header")
		return
	}

	const prefix = "Bearer "
	
	if !strings.HasPrefix(authHeader, prefix){
		cfg.respondWithError(w, 401, "invalid authorization header")
		return
	}
	authToken := strings.TrimPrefix(authHeader, prefix)
	
	usrID, err := auth.ValidateJWT(authToken, cfg.jwtSecret)
	if err != nil{
		cfg.respondWithError(w, 401, "Unauthorized")
		return
	}
	
	//update the user credentials
	newUsr, err := cfg.db.UpdateCredentials(r.Context(), database.UpdateCredentialsParams{
		Email: params.Email,
		HashedPassword: hashedPassword,
		ID: usrID,
	})

	if err != nil{
		cfg.respondWithError(w, 500, "Error Updating Credentials")
		return
	}

	//now return the new user resource

	resp := okResponse{ID: newUsr.ID, CreatedAt: newUsr.CreatedAt, UpdatedAt: newUsr.UpdatedAt, Email: newUsr.Email, IsChirpyRed: newUsr.IsChirpyRed.Bool}

	cfg.respondWithJSON(w, 200, resp)


}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request){
	//grab access token
	//then provide new email and password
	
	//Have to check that the user has a bearerID that is valid 
	//get the bearer token from the header
	//then validate that bearer token

	parsedUUID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Fatalf("failed to parse UUID string: %v", err)
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == ""{
		cfg.respondWithError(w, 401, "missing authorization header")
		return
	}

	const prefix = "Bearer "
	
	if !strings.HasPrefix(authHeader, prefix){
		cfg.respondWithError(w, 401, "invalid authorization header")
		return
	}
	authToken := strings.TrimPrefix(authHeader, prefix)
	
	usrID, err := auth.ValidateJWT(authToken, cfg.jwtSecret)
	if err != nil{
		cfg.respondWithError(w, 401, "Unauthorized")
		return
	}
	
	chirp, err := cfg.db.GetChirpByID(r.Context(), parsedUUID)
	if err != nil{
		cfg.respondWithError(w, 404, "Resource not found")
		return	
	}

	if usrID != chirp.UserID{
		cfg.respondWithError(w, 403, "User with Token ID does not match Chirp User ID")
		return
	}

	err = cfg.db.DeleteChirpByID(r.Context(), usrID)
	if err != nil{
		cfg.respondWithError(w, 500, "Ther was an issue deleting the chirp")
		return	
	}
	//now return the new user resource
	w.WriteHeader(204)
}

func(cfg *apiConfig) upgradeUser(w http.ResponseWriter, r *http.Request){
	
	type parameters struct{
		Event string `json:"event"`
		Data struct{
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	
	authHead, err := auth.GetAPIKEY(r.Header)
	log.Printf("This is the authhead %v", authHead)
	log.Printf("This is the apiKey %v", cfg.polkaKey)
	if authHead != cfg.polkaKey{
		cfg.respondWithError(w, 401, "API key not valid")
		return
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil{
		log.Printf("Error marshalling data %s", err)
		w.WriteHeader(500)
		return
	}

	//check that all fields contain relevant info
	if params.Event != "user.upgraded"{
		cfg.respondWithError(w, 404, "Not valid event")
		return
	}

	_, err = cfg.db.UpgradeUser(r.Context(), params.Data.UserID)
	if err != nil{
		cfg.respondWithError(w, 404, "User not found")
		return
	}

	w.WriteHeader(204) // Sets the status code
	fmt.Fprintf(w, "204 User upgraded") // Writes the body content
	return


	
}	

