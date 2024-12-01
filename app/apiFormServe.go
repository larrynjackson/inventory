package app

import (
	"fmt"
	"github/lnj/inventory/model"
	"html/template"
	"log"
	"net/http"
	"os"
)

func (s *Server) apiIndex(w http.ResponseWriter, r *http.Request) {
	simpleFormHelper(w, "home.html", "Welcome")
}

func (s *Server) apiGetLoginForm(w http.ResponseWriter, r *http.Request) {
	simpleFormHelper(w, "login.html", "")
}

func (s *Server) apiGetAdminForm(w http.ResponseWriter, r *http.Request) {
	s.simpleLoginFormHelper(w, r, "admin.html")
}

func (s *Server) apiGetMenuForm(w http.ResponseWriter, r *http.Request) {
	s.simpleLoginFormHelper(w, r, "menu.html")
}

func (s *Server) apiIgnore(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) apiServeFile(w http.ResponseWriter, r *http.Request) {
	var token = getCookie(r)

	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	s.sendFileHelper(w, r)
}

func (s *Server) apiServeFlagFile(w http.ResponseWriter, r *http.Request) {
	s.sendFileHelper(w, r)
}

func (s *Server) apiGetFileForm(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var token = getCookie(r)

	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	itemSelect := r.URL.Query().Get("itemSelect")
	if itemSelect == "" {
		itemSelect = "Choose an Item"
	}
	fmt.Println("fileform item:", itemSelect)
	tKey := s.getTKeyHelper(token)
	message, ok := s.fillItemsHelper(&items, itemSelect, tKey)
	if !ok {
		items.Message = message
		executeTemplateReturn(w, &items, "menu.html")
		return
	}
	items.Message = ""
	executeTemplateReturn(w, &items, "file.html")
}

func (s *Server) apiGetCreateItemForm(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var token = getCookie(r)

	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	items.Message = ""
	executeTemplateReturn(w, &items, "createitem.html")
}

func (s *Server) apiGetEditItemForm(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var token = getCookie(r)

	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	itemSelect := r.URL.Query().Get("itemSelect")
	tKey := s.getTKeyHelper(token)
	message, ok := s.fillItemsHelper(&items, itemSelect, tKey)
	if !ok {
		items.Message = message
		executeTemplateReturn(w, &items, "menu.html")
		return
	}
	items.Message = ""
	executeTemplateReturn(w, &items, "edititem.html")
}

func (s *Server) apiGetShowItemsForm(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var token = getCookie(r)

	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	itemSelect := r.URL.Query().Get("itemSelect")
	tKey := s.getTKeyHelper(token)
	message, ok := s.fillItemsHelper(&items, itemSelect, tKey)
	if !ok {
		items.Message = message
		executeTemplateReturn(w, &items, "menu.html")
		return
	}
	if itemSelect == "" {
		itemSelect = items.ItemList[0].Key
	}
	files, err := os.ReadDir("./images/" + itemSelect)
	if err != nil {
		log.Fatal(err)
	}
	if len(files) > 0 {
		currentFile := "http://localhost:8000/api/v1/getfile/images/" + itemSelect + "/" + files[0].Name()
		items.Pictures.DisplayFile = template.URL(currentFile)

		items.Pictures.DisableLeft = "disabled"
		if len(files) == 1 {
			items.Pictures.DisableRight = "disabled"
		} else {
			items.Pictures.DisableRight = ""
		}
		items.Pictures.CurrentFile = files[0].Name()
	} else {
		items.Pictures.DisableLeft = "disabled"
		items.Pictures.DisableRight = "disabled"
		items.Pictures.CurrentFile = "NoFiles"
	}
	items.Message = ""
	executeTemplateReturn(w, &items, "showitems.html")
}
