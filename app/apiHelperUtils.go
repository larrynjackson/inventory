package app

import (
	"fmt"
	"github/lnj/inventory/model"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func (s *Server) sendFileHelper(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileName := vars["filename"]
	directory := vars["directory"]
	if directory == "" {
		directory = "navbar"
	}
	buf, err := os.ReadFile("images/" + directory + "/" + fileName)
	if err != nil {
		panic(err)
	}
	w.Write(buf)
}

func (s *Server) simpleLoginFormHelper(w http.ResponseWriter, r *http.Request, htmlFormFile string) {
	var items model.Items
	var token = getCookie(r)

	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	items.Message = ""
	executeTemplateReturn(w, &items, htmlFormFile)
}

func simpleFormHelper(w http.ResponseWriter, htmlFormFile, message string) {
	var items model.Items

	items.Message = message
	executeTemplateReturn(w, &items, htmlFormFile)
}

func (s *Server) isLoggedIn(token string) bool {
	mapFrame, ok := s.controlMap.GetMap(token)
	if !ok {
		fmt.Println("no mapFrame for token:", token)
		return false
	}
	nextAction := mapFrame.context.nextAction
	mapFrame.createTime = time.Now()
	mapFrame.timeToLive = time.Duration(30 * time.Minute)
	s.controlMap.PushMap(token, mapFrame)
	return nextAction == ClientActionInventory
}

func executeTemplateReturn(w http.ResponseWriter, items *model.Items, htmlFile string) {
	items.MenuImageFile = template.URL(LocalFreeFlagURL)
	tmpl, err := template.ParseFiles("views/" + htmlFile)
	if err != nil {
		log.Printf("ERROR: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, items)
}

func (s *Server) getTKeyHelper(token string) string {
	mapFrame, _ := s.controlMap.GetMap(token)
	return mapFrame.context.tKey
}

func (s *Server) fillItemsHelper(items *model.Items, itemSelect, tKey string) (string, bool) {
	var itemList []model.Item
	keyMapArray := s.Storage.SelectItems(tKey)
	ok := true
	if len(keyMapArray) == 0 {
		return "No Items Found.", !ok
	}
	for i := 0; i < len(keyMapArray); i++ {
		var item model.Item
		if itemSelect == "" && i == 0 {
			item.Selected = "selected"
			items.SelectedItemName = keyMapArray[i]["name"]
			items.SelectedItemDescription = keyMapArray[i]["description"]
			items.SelectedItemValue = keyMapArray[i]["value"]
			items.SelectedItemPurchaseDate = keyMapArray[i]["purchasedate"]
			items.SelectedItemSerialNum = keyMapArray[i]["serialnum"]
		} else if itemSelect == keyMapArray[i]["dKey"] {
			item.Selected = "selected"
			items.SelectedItemName = keyMapArray[i]["name"]
			items.SelectedItemDescription = keyMapArray[i]["description"]
			items.SelectedItemValue = keyMapArray[i]["value"]
			items.SelectedItemPurchaseDate = keyMapArray[i]["purchasedate"]
			items.SelectedItemSerialNum = keyMapArray[i]["serialnum"]
		} else {
			item.Selected = ""
		}
		item.Key = keyMapArray[i]["dKey"]
		item.Name = keyMapArray[i]["name"]
		itemList = append(itemList, item)
	}
	items.ItemList = itemList
	return "", ok
}

func copyFileHelper(src, dest string) string {

	inFile, err := os.Open(src)
	if err != nil {
		return "Can not open input file."
	}

	defer func() {
		if err := inFile.Close(); err != nil {
			panic(err)
		}
	}()

	outFile, err := os.Create(dest)
	if err != nil {
		return "Can not create output file"
	}

	defer func() {
		if err := outFile.Close(); err != nil {
			panic(err)
		}
	}()

	buf := make([]byte, 1024)
	for {

		n, err := inFile.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		if _, err := outFile.Write(buf[:n]); err != nil {
			panic(err)
		}
	}
	return "File copy complete."
}
