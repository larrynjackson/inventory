package app

import (
	"github/lnj/inventory/model"
	"net/http"
)

func (s *Server) apiLogin(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var targetHtmlFile string = "menu.html"

	message, ok := s.handleLogin(w, r)
	if !ok {
		targetHtmlFile = "login.html"
	}

	items.Message = message
	executeTemplateReturn(w, &items, targetHtmlFile)
}

func (s *Server) apiLogout(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var targetHtmlFile string = "home.html"

	var token = getCookie(r)
	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	message := s.handleLogout(w, r)
	items.Message = message
	executeTemplateReturn(w, &items, targetHtmlFile)
}

func (s *Server) apiDeleteUser(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var targetHtmlFile string = "home.html"

	var token = getCookie(r)
	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	message, ok := s.handleDeleteUser(w, r)
	if !ok {
		targetHtmlFile = "admin.html"
	}
	items.Message = message
	executeTemplateReturn(w, &items, targetHtmlFile)

}

func (s *Server) apiAddUser(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var targetHtmlFile string = "matchcode.html"

	message, ok := s.handleAddUser(w, r)
	if !ok {
		targetHtmlFile = "login.html"
	}

	items.Message = message
	executeTemplateReturn(w, &items, targetHtmlFile)
}

func (s *Server) apiMatchCode(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var targetHtmlFile string = "login.html"

	message, _ := s.handleMatchCode(w, r)

	items.Message = message
	executeTemplateReturn(w, &items, targetHtmlFile)
}

func (s *Server) apiChangePassword(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var targetHtmlFile string = "menu.html"

	var token = getCookie(r)
	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	message, ok := s.handleChangePassword(w, r)
	if !ok {
		targetHtmlFile = "admin.html"
	}
	items.Message = message
	executeTemplateReturn(w, &items, targetHtmlFile)
}

func (s *Server) apiResetPassword(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var targetHtmlFile string = "login.html"

	message, _ := s.handleResetPassword(w, r)

	items.Message = message
	executeTemplateReturn(w, &items, targetHtmlFile)
}
