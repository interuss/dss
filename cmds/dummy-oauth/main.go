// ?grant_type=client_credentials&scope={}&intended_audience={}
package main

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/dgrijalva/jwt-go"
)

var (
	address = flag.String("addr", ":8085", "address")
	keyFile = flag.String("private_key_file", "config/test-certs/oauth.key", "oauth private key file")
)

// TODO(steeling): add other parameters so we can test expired tokens, invalid tokens, etc.
func createGetTokenHandler(privateKey *rsa.PrivateKey) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestBytes, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Panic(err)
		} else {
			log.Println(string(requestBytes))
		}

		w.Header().Set("Content-Type", "application/json")
		params := r.URL.Query()

		var (
			aud  string   = ""
			auds []string = params["intended_audience"]
		)
		if len(auds) == 1 {
			aud = auds[0]
		}

		var (
			scope  string   = ""
			scopes []string = params["scope"]
		)
		if len(scopes) == 1 {
			scope = scopes[0]
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"aud":   aud,
			"scope": scope,
			"sub":   "fake-user",
		})

		// Sign and get the complete encoded token as a string using the secret
		// Ignore the error, it will fail the test anyways if it is not nil.
		tokenString, err := token.SignedString(privateKey)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		encodeError := json.NewEncoder(w).Encode(map[string]string{
			"access_token": tokenString,
		})
		if encodeError != nil {
			log.Panic(encodeError)
		}
	})
}

func readPrivateKey() (*rsa.PrivateKey, error) {
	bytes, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPrivateKeyFromPEM(bytes)
}

func main() {
	flag.Parse()
	privateKey, err := readPrivateKey()
	if err != nil {
		log.Panic(err)
	}
	http.Handle("/token", createGetTokenHandler(privateKey))
	log.Panic(http.ListenAndServe(*address, nil))
}
