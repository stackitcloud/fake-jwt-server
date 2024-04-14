package server

import (
	"crypto/rsa"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

//go:embed key.pem
var embeddedPrivateKey string

type TokenResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type FakeJWTServer struct {
	privateKey   *rsa.PrivateKey
	jsonWebKey   jwk.Key
	publicKeySet jwk.Set
	config       Config
}

func (f *FakeJWTServer) TokenHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		log.Printf("handle token request")
		writer.Header().Add("Content-Type", "application/json")

		token, err := f.createJsonWebToken()
		if err != nil {
			log.Printf("failed to create json web token: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)

			return
		}

		responseData, err := json.MarshalIndent(TokenResponse{
			Message:     "Successfully created token",
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int(f.config.Expires.Seconds()),
		}, "", "  ")
		if err != nil {
			log.Printf("failed to marshal response: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)

			return
		}
		if _, err := writer.Write(responseData); err != nil {
			log.Printf("failed to write response: %v", err)

			return
		}
	default:
		log.Printf("unexpected method for token handler. expected post.")

		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func (f *FakeJWTServer) JwksHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		log.Printf("handle jwks request")
		writer.Header().Add("Content-Type", "application/json")

		responseData, err := json.MarshalIndent(f.publicKeySet, "", " ")
		if err != nil {
			log.Printf("failed to marshal public key set: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)

			return
		}
		if _, err := writer.Write(responseData); err != nil {
			log.Printf("failed to write response: %v", err)

			return
		}
	default:
		log.Printf("unexpected method for jwks handler. expected get.")
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func loadPrivateKey() (*rsa.PrivateKey, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(embeddedPrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}

func createJsonWebKey(privateKey *rsa.PrivateKey) (jwk.Key, error) {
	key, err := jwk.FromRaw(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWK from private key: %w", err)
	}
	if err := key.Set(jwk.KeyUsageKey, jwk.ForSignature); err != nil {
		return nil, fmt.Errorf("failed to set key usage: %w", err)
	}
	if err := key.Set(jwk.AlgorithmKey, jwa.RS512); err != nil {
		return nil, fmt.Errorf("failed to set key algorithm: %w", err)
	}
	if err := key.Set(jwk.KeyIDKey, "683a2fae-2be1-4fd7-85f5-0e538e627c22"); err != nil {
		return nil, fmt.Errorf("failed to set key id: %w", err)
	}

	return key, nil
}

func createPublicKeySet(jsonWebKey jwk.Key) (jwk.Set, error) {
	keySet := jwk.NewSet()
	if err := keySet.AddKey(jsonWebKey); err != nil {
		return nil, fmt.Errorf("failed to add JWK to key set: %w", err)
	}

	publicKeySet, err := jwk.PublicSetOf(keySet)
	if err != nil {
		return nil, fmt.Errorf("failed to create public key set: %w", err)
	}

	return publicKeySet, nil
}

type FakeClaims struct {
	jwt.RegisteredClaims
	GrantType string `json:"grant_type,omitempty"`
	Email     string `json:"email,omitempty"`
}

func (f *FakeJWTServer) createJsonWebToken() (string, error) {
	claims := FakeClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    f.config.Issuer,
			Subject:   f.config.Subject,
			Audience:  jwt.ClaimStrings{f.config.Audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(f.config.Expires)),
			NotBefore: jwt.NewNumericDate(time.Now().AddDate(0, 0, -1)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        f.config.ID,
		},
		Email:     f.config.Email,
		GrantType: f.config.GrandType,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	token.Header[jwk.KeyIDKey] = f.jsonWebKey.KeyID()
	signedToken, err := token.SignedString(f.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

type Config struct {
	Issuer    string
	Subject   string
	Audience  string
	ID        string
	Port      int
	Expires   time.Duration
	Email     string
	GrandType string
}

func NewFakeJwtServer() *FakeJWTServer {
	privateKey, err := loadPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	jsonWebKey, err := createJsonWebKey(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKeySet, err := createPublicKeySet(jsonWebKey)
	if err != nil {
		log.Fatal(err)
	}

	return &FakeJWTServer{
		privateKey:   privateKey,
		jsonWebKey:   jsonWebKey,
		publicKeySet: publicKeySet,
		config: Config{
			Issuer:    "test",
			Subject:   "test",
			Audience:  "test",
			ID:        "test",
			Port:      8008,
			Expires:   24 * 365 * 100 * time.Hour,
			Email:     "test@example.com",
			GrandType: "client_credentials",
		},
	}
}

func (f *FakeJWTServer) WithAudience(audience string) *FakeJWTServer {
	f.config.Audience = audience

	return f
}

func (f *FakeJWTServer) WithIssuer(issuer string) *FakeJWTServer {
	f.config.Issuer = issuer

	return f
}

func (f *FakeJWTServer) WithSubject(subject string) *FakeJWTServer {
	f.config.Subject = subject

	return f
}

func (f *FakeJWTServer) WithID(id string) *FakeJWTServer {
	f.config.ID = id

	return f
}

func (f *FakeJWTServer) WithPort(port int) *FakeJWTServer {
	f.config.Port = port

	return f
}

func (f *FakeJWTServer) WithExpires(expires time.Duration) *FakeJWTServer {
	f.config.Expires = expires

	return f
}

func (f *FakeJWTServer) WithEmail(email string) *FakeJWTServer {
	f.config.Email = email

	return f
}

func (f *FakeJWTServer) WithGrandType(grantType string) *FakeJWTServer {
	f.config.GrandType = grantType

	return f
}

func (f *FakeJWTServer) Serve() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", f.TokenHandler)
	mux.HandleFunc("/jwks", f.JwksHandler)
	mux.HandleFunc("/.well-known/jwks.json", f.JwksHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", f.config.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return server.ListenAndServe()
}
