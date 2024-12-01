package app

import "github.com/gorilla/mux"

func versionedApiRoutes(versionRouter *mux.Router, s *Server) {

	versionRouter.HandleFunc("/loginform", s.apiGetLoginForm).Methods("GET")
	versionRouter.HandleFunc("/fileform", s.apiGetFileForm).Methods("GET")
	versionRouter.HandleFunc("/adminform", s.apiGetAdminForm).Methods("GET")
	versionRouter.HandleFunc("/createitemform", s.apiGetCreateItemForm).Methods("GET")
	versionRouter.HandleFunc("/edititemform", s.apiGetEditItemForm).Methods("GET")
	versionRouter.HandleFunc("/menuform", s.apiGetMenuForm).Methods("GET")
	versionRouter.HandleFunc("/showitemsform", s.apiGetShowItemsForm).Methods("GET")
	versionRouter.HandleFunc("/logout", s.apiLogout).Methods("GET")

	versionRouter.HandleFunc("/getfile/images/{directory}/{filename}", s.apiServeFile).Methods("GET")
	versionRouter.HandleFunc("/getfile/{filename}", s.apiServeFlagFile).Methods("GET")

	versionRouter.HandleFunc("/nextpic", s.apiGetNextPicture).Methods("POST")
	versionRouter.HandleFunc("/copyfile", s.apiCopyFile).Methods("POST")
	versionRouter.HandleFunc("/createitem", s.apiCreateItem).Methods("POST")
	versionRouter.HandleFunc("/edititem", s.apiEditItem).Methods("POST")
	versionRouter.HandleFunc("/deleteitem", s.apiDeleteItem).Methods("POST")

	versionRouter.HandleFunc("/delete", s.apiDeleteUser).Methods("POST")
	versionRouter.HandleFunc("/auth", s.apiLogin).Methods("POST")
	versionRouter.HandleFunc("/code", s.apiMatchCode).Methods("POST")
	versionRouter.HandleFunc("/add/user", s.apiAddUser).Methods("POST")
	versionRouter.HandleFunc("/changepassword", s.apiChangePassword).Methods("POST")
	versionRouter.HandleFunc("/resetpassword", s.apiResetPassword).Methods("POST")

	versionRouter.HandleFunc("/pokemon/start", s.handlePokemonStart).Methods("GET")
	versionRouter.HandleFunc("/pokemon/poke", s.handlePokemonPoke).Methods("POST")
}
