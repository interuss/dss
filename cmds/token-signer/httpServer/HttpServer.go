package httpServer

import (
	"encoding/json"
	"log"
	"net/http"
	"token-signer/config"
	"token-signer/tokenSigner"
)

type HttpServer struct {
}

type RequestParser interface {
	ParseRequest(r *http.Request) (tokenSigner.CreateSignedTokenRequest, error)
}

func (h HttpServer) Serve(signer tokenSigner.TokenSigner) {

	conf := config.GetGlobalConfig()
	var parser RequestParser = HttpRequestParser{}

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Token requested with parameters")
		log.Println(r)

		if r.Method != "GET" {
			http.Error(w, "Invalid request method", 405)
			return
		}
		request, err := parser.ParseRequest(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		response, err := signer.CreateSignedToken(request)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Error", 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Error", 500)
			return
		}

		return
	})
	log.Print("Server started on port :" + conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, nil))
}
