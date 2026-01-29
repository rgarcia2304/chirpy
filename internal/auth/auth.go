package auth 

import(
	"github.com/alexedwards/argon2id"
	"errors"
	"net/http"
	"strings"
)

func HashPassword(password string) (string, error){
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil{
		return "", errors.New("There was an issue hashing the password")
	}

	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error){
	//compare the pasword entered in the http request with the one that was hashed in db
	same, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil{
		return false, errors.New("There was an error when comparing passwords")
	}

	if same{
		return true, nil
	}else{
		return false, nil
	}
}

func GetBearerToken(headers http.Header) (string, error){
	
	val := headers.Get("Authorization")
	if val == ""{
		return "", errors.New("There was not Authorization string provided")
	}

	//strip the prefix and remove whitespace
	trimmed := strings.Trim(val, "Bearer")
	trimmed = strings.TrimSpace(trimmed)
	return trimmed, nil
}	
