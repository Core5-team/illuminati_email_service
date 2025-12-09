package routes

import (
	"net/http"
)

func SetupAPI(){
	http.HandleFunc("/send_letter", ls.PostLetter)
}

