// ?grant_type=client_credentials&scope={}&intended_audience={}
package main

import (
  "crypto/rsa"
  "encoding/json"
  "flag"
  "io/ioutil"
  "log"
  "net/http"

  "github.com/dgrijalva/jwt-go"
)

var (
  address    = flag.String("addr", ":8085", "address")
  keyFile    = flag.String("private_key_file", "config/test-certs/oauth.key", "oauth private key file")
  privateKey *rsa.PrivateKey
)

// TODO(steeling): add other parameters so we can test expired tokens, invalid tokens, etc.
func getToken(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  params := r.URL.Query()
  token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
    "aud":   params["intended_audience"],
    "scope": params["scope"],
  })

  // Sign and get the complete encoded token as a string using the secret
  // Ignore the error, it will fail the test anyways if it is not nil.
  tokenString, err := token.SignedString(privateKey)
  if err != nil {
    http.Error(w, err.Error(), 500)
    return
  }

  json.NewEncoder(w).Encode(map[string]string{
    "access_token": tokenString,
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
  var err error
  privateKey, err = readPrivateKey()
  if err != nil {
    log.Fatal(err)
  }
  http.HandleFunc("/token", getToken)
  log.Fatal(http.ListenAndServe(*address, nil))

}
