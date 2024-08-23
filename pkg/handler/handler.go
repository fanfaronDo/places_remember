package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	types "main/internal/domain/place"
	db "main/pkg/db"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type Handler struct{}

type PlacesTemplatePagination struct {
	Name     string        `json:"name"`
	Total    int           `json:"total"`
	Places   []types.Place `json:"places"`
	PrevPage int           `json:"prev_page"`
	NextPage int           `json:"next_page"`
	LastPage int           `json:"last_page"`
}

const (
	Limit = 10
)

func (h *Handler) InitRoutes() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", h.pages)
	router.HandleFunc("/api/get_jwt", h.getTokenHandler)
	routerApi := router.PathPrefix("/api").Subrouter()
	routerApi.Methods("GET").Path("/places").HandlerFunc(h.apiPage)
	routerApi.Methods("GET").Path("/recommend").HandlerFunc(h.apiRecommend)
	routerApi.Use(h.jwtMiddleware)

	return router
}

func (h *Handler) pages(w http.ResponseWriter, r *http.Request) {
	var templatePagination PlacesTemplatePagination

	id, err := strconv.Atoi(r.URL.Query().Get("page"))

	if err != nil || id < 1 {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templatePagination, err = h.setPlacesPagination(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dir, _ := os.Getwd()
	path := filepath.Join(dir, "web", "template", "index.html")
	file, err := ioutil.ReadFile(path)

	tmp, err := template.New("places").Parse(string(file))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmp.Execute(w, templatePagination)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) apiPage(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || id < 1 {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templatePagination, err := h.setPlacesPagination(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(templatePagination)
	w.Write(data)
}

func (h *Handler) apiRecommend(w http.ResponseWriter, r *http.Request) {
	lat, err1 := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, err2 := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)

	if err1 != nil || err2 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
		return
	}

	places, err := db.NewPlaceStore().GetRecommend(lat, lon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(places)

	w.Write(data)
}

func (h *Handler) setPlacesPagination(id int) (PlacesTemplatePagination, error) {

	tmpPagination := PlacesTemplatePagination{}
	tmpPagination.Name = "Places"
	offset := id * Limit

	places, total, err := db.NewPlaceStore().GetPlaces(Limit, offset)

	if err != nil {
		return tmpPagination, err
	}

	lastPage := total / Limit

	tmpPagination.Total = total
	tmpPagination.Places = places
	tmpPagination.LastPage = lastPage

	if id == 1 {
		tmpPagination.PrevPage = 0
		tmpPagination.NextPage = id + 1
	} else if total-offset < Limit {
		tmpPagination.PrevPage = id - 1
		tmpPagination.NextPage = 0
	} else {
		tmpPagination.PrevPage = id - 1
		tmpPagination.NextPage = id + 1
	}

	return tmpPagination, nil
}

func (h *Handler) getTokenHandler(w http.ResponseWriter, r *http.Request) {

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("token"))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": signedToken,
	})
}

func (h *Handler) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(""), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
