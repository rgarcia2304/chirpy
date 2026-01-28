package main 
import(
	"strings"
)

func profaneCheck(msg string) (string, bool){
	
	convertedMsg:= strings.ToLower(msg)
	splitMsg := strings.Split(convertedMsg, " ")
	profane := false
	
	newMsg := "" 
	for _, word := range splitMsg{
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			newMsg += "****" + " " 
			profane = true
		}else{
			newMsg += word + " "
		}
	}

	return newMsg, profane
}
