package app

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type Pokemon struct {
	Name           string  `json:"name"`
	BaseExperience int     `json:"base_experience"`
	Height         int     `json:"height"`
	Weight         int     `json:"weight"`
	Sprites        Sprites `json:"sprites"`
	Types          []Type  `json:"types"`
}

type Sprites struct {
	Other        OtherSprites `json:"other"`
	FrontDefault string       `json:"front_default"`
}

type OtherSprites struct {
	Showdown Showdown `json:"showdown"`
}

type Showdown struct {
	FrontDefault string `json:"front_default"`
}

type Type struct {
	Slot       int        `json:"slot"`
	TypeDetail TypeDetail `json:"type"`
}

type TypeDetail struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

var t = template.Must(template.ParseGlob("./views/pokemon/*.html"))

func (s *Server) handlePokemonStart(w http.ResponseWriter, r *http.Request) {

	fmt.Println("pokemon/start")

	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon/pichu")
	if err != nil {
		http.Error(w, "unable to grap the Pokemon data", http.StatusInternalServerError)
	}

	data := Pokemon{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		http.Error(w, "unable to parse Pokemon data", http.StatusInternalServerError)
	}

	if err := t.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
}

func (s *Server) handlePokemonPoke(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse for", http.StatusInternalServerError)
	}

	fmt.Println("pokemon/poke")

	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + strings.ToLower(r.FormValue("pokemon")))
	if err != nil {
		http.Error(w, "unable to grap the Pokemon data", http.StatusInternalServerError)
	}
	data := Pokemon{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		http.Error(w, "unable to parse Pokemon data", http.StatusInternalServerError)
	}
	fmt.Println("data:", data)
	if data.Name != "" {
		if err := t.ExecuteTemplate(w, "response.html", data); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	} else {
		if err := t.ExecuteTemplate(w, "error.html", nil); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	}

}
