package app

import (
	"fmt"
	"github/lnj/inventory/model"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func (s *Server) apiGetNextPicture(w http.ResponseWriter, r *http.Request) {
	var items model.Items

	itemSelect := r.PostFormValue("selectedItem")
	currentFile := r.PostFormValue("currentFile")
	fileDirection := r.PostFormValue("fileDirection")

	files, err := os.ReadDir("./images/" + itemSelect)
	if err != nil {
		log.Fatal(err)
	}
	var numFiles = len(files)
	var currentIdx int
	for idx := 0; idx < numFiles; idx++ {
		if files[idx].Name() == currentFile {
			currentIdx = idx
			break
		}
	}
	if currentIdx > 0 && fileDirection == "prev" {
		currentIdx--
	}
	if currentIdx < (numFiles-1) && fileDirection == "next" {
		currentIdx++
	}
	currentFile = files[currentIdx].Name()
	if currentIdx > 0 {
		items.Pictures.DisableLeft = ""
	} else {
		items.Pictures.DisableLeft = "disabled"
	}
	if currentIdx < (numFiles - 1) {
		items.Pictures.DisableRight = ""
	} else {
		items.Pictures.DisableRight = "disabled"
	}
	items.Pictures.CurrentFile = currentFile
	fileURL := "http://localhost:8000/api/v1/getfile/images/" + itemSelect + "/" + currentFile
	items.Pictures.DisplayFile = template.URL(fileURL)
	executeTemplateReturn(w, &items, "nextpic.html")
}

func (s *Server) apiEditItem(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var token = getCookie(r)

	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	itemSelect := r.PostFormValue("selectedItem")
	name := r.PostFormValue("name")
	description := r.PostFormValue("description")
	value := r.PostFormValue("value")
	purchaseDate := r.PostFormValue("purchasedate")
	serialNum := r.PostFormValue("serialnum")

	name = strings.ReplaceAll(name, "'", "''")
	description = strings.ReplaceAll(description, "'", "''")
	value = strings.ReplaceAll(value, "'", "''")
	purchaseDate = strings.ReplaceAll(purchaseDate, "'", "''")
	serialNum = strings.ReplaceAll(serialNum, "'", "''")
	description = strings.TrimSpace(description)

	tKey := s.getTKeyHelper(token)
	s.Storage.UpdateItem(itemSelect, name, description, value, purchaseDate, serialNum)
	fmt.Println("editItem:", itemSelect)
	message, ok := s.fillItemsHelper(&items, itemSelect, tKey)
	if !ok {
		items.Message = message
		executeTemplateReturn(w, &items, "menu.html")
		return
	}
	items.Message = "Item Saved"
	executeTemplateReturn(w, &items, "edititem.html")
}

func (s *Server) apiDeleteItem(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var token = getCookie(r)

	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	itemSelect := r.PostFormValue("selectedItem")
	err := os.Chdir("./images")
	if err != nil {
		panic(err)
	}
	fmt.Println("remove:", itemSelect)
	err = os.RemoveAll("./" + itemSelect)
	if err != nil {
		panic(err)
	}
	err = os.Chdir("..")
	if err != nil {
		panic(err)
	}
	s.Storage.DeleteItem(itemSelect)
	tKey := s.getTKeyHelper(token)
	message, ok := s.fillItemsHelper(&items, itemSelect, tKey)
	if !ok {
		items.Message = message
		executeTemplateReturn(w, &items, "menu.html")
		return
	}
	items.Message = "Item Saved"
	executeTemplateReturn(w, &items, "edititem.html")
}

func (s *Server) apiCreateItem(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var token = getCookie(r)

	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	name := r.PostFormValue("name")
	description := r.PostFormValue("description")
	value := r.PostFormValue("value")
	purchaseDate := r.PostFormValue("purchasedate")
	serialNum := r.PostFormValue("serialnum")

	name = strings.ReplaceAll(name, "'", "''")
	description = strings.ReplaceAll(description, "'", "''")
	value = strings.ReplaceAll(value, "'", "''")
	purchaseDate = strings.ReplaceAll(purchaseDate, "'", "''")
	serialNum = strings.ReplaceAll(serialNum, "'", "''")
	description = strings.TrimSpace(description)

	tKey := s.getTKeyHelper(token)
	dKey := GenerateRandomKey(4)
	fmt.Println("dKey:", dKey)
	s.Storage.InsertItem(tKey, dKey, name, description, value, purchaseDate, serialNum)
	dirPath := "./images/" + dKey
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		panic(err)
	}
	items.Message = "Item Inserted"
	executeTemplateReturn(w, &items, "createitem.html")
}

func (s *Server) apiCopyFile(w http.ResponseWriter, r *http.Request) {
	var items model.Items
	var token = getCookie(r)

	if !s.isLoggedIn(token) {
		w.Write([]byte("<h1>Good Bye</h1>"))
		return
	}
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("copyfile")
	if err != nil {
		items.Message = "Read file failed."
		executeTemplateReturn(w, &items, "file.html")
	}
	defer file.Close()
	fmt.Printf("upload file: %+v\n", handler.Filename)
	fmt.Printf("file size: %+v\n", handler.Size)

	itemSelect := r.PostFormValue("selectedItem")

	var destFile string = "./images/" + itemSelect + "/" + handler.Filename

	tKey := s.getTKeyHelper(token)

	tempFile, err := os.Create(destFile)
	if err != nil {
		items.Message = "File copy failed."
		executeTemplateReturn(w, &items, "file.html")
	}
	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		items.Message = "Read file failed."
		executeTemplateReturn(w, &items, "file.html")
	}

	tempFile.Write(fileBytes)

	PrintCollisionCount()
	PrintKeyMapSize()

	message, ok := s.fillItemsHelper(&items, itemSelect, tKey)
	if !ok {
		items.Message = message
		executeTemplateReturn(w, &items, "menu.html")
		return
	}

	items.Message = "Upload complete"
	executeTemplateReturn(w, &items, "file.html")
}
