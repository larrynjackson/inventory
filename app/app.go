package app

import (
	"crypto/sha256"
	"encoding/hex"
	"github/lnj/inventory/data"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/rand"
)

type Server struct {
	Listen     string
	router     *mux.Router
	Storage    data.StorageReadWrite
	controlMap ControlMap
}

func NewServer(storageRW data.StorageReadWrite, listenAddress string) *Server {
	return &Server{
		Listen:  listenAddress,
		router:  mux.NewRouter(),
		Storage: storageRW,
	}
}

type MapFrame struct {
	token      string
	createTime time.Time
	timeToLive time.Duration
	context    UserContext
}

type UserContext struct {
	userId     string
	tKey       string
	currAction string
	nextAction string
	password   string
	machineId  string
	passCode   string
}

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var mapLock sync.Mutex

type ControlMap map[string]MapFrame

func (m ControlMap) PushMap(token string, mapFrame MapFrame) {
	mapLock.Lock()
	defer mapLock.Unlock()
	m[token] = mapFrame
}

func (m ControlMap) GetMap(token string) (MapFrame, bool) {
	mapLock.Lock()
	defer mapLock.Unlock()
	smf, ok := m[token]
	return smf, ok
}
func (m ControlMap) PopMap(token string) bool {
	mapLock.Lock()
	defer mapLock.Unlock()
	_, ok := m[token]
	if ok {
		delete(m, token)
	}
	return ok
}

func GenerateToken(machineId string) string {

	mapLock.Lock()
	defer mapLock.Unlock()

	tokenString := time.Now().String() + machineId
	tokenHex := sha256.Sum256([]byte(tokenString))
	token := hex.EncodeToString(tokenHex[:])

	return token
}

var keyMap = make(map[string]string)
var collisionCount = 0

var lock sync.Mutex

func GenerateRandomKey(length int) string {
	lock.Lock()
	defer lock.Unlock()

	strArr := []rune(charset)
	var key string
	var done bool = false
	for !done {
		key = ""
		for j := 0; j < length; j++ {
			key += string(strArr[rand.Intn(len(charset))])
		}
		hasKey := keyMap[key]
		if hasKey == key {
			collisionCount++
			log.Info().Msgf("collision:'%s'", key)
		} else {
			keyMap[key] = key
			done = true
		}
	}
	return key
}

func PrintCollisionCount() {
	log.Info().Msgf("Collision count:'%d'", collisionCount)
}

func PrintKeyMapSize() {
	log.Info().Msgf("keyMap count:'%d'", len(keyMap))
}

func (s *Server) CleanStackMap() {
	for {
		time.Sleep(15 * time.Minute)
		log.Info().Msg("Running 15 minute CleanStackMap function.")

		mapLock.Lock()

		log.Info().Msgf("map count before:'%d'", len(s.controlMap))
		for _, mf := range s.controlMap {
			var now = time.Now()
			if now.After(mf.createTime.Add(mf.timeToLive)) {
				delete(s.controlMap, mf.token)
				log.Info().Msgf("map count after:'%d'", len(s.controlMap))
			}
		}
		mapLock.Unlock()
	}
}

func (s *Server) Run() {
	credentials := handlers.AllowCredentials()
	methods := handlers.AllowedMethods([]string{"POST", "DELETE", "GET"})
	origins := handlers.AllowedOrigins([]string{"http://localhost:8000"})

	s.controlMap = make(ControlMap)

	// Handle basic root paths
	s.router.HandleFunc("/", s.apiIndex)
	s.router.HandleFunc("/favicon.ico", s.apiIgnore)

	api := s.router.PathPrefix("/api").Subrouter()
	v1 := api.PathPrefix("/v1").Subrouter()
	versionedApiRoutes(v1, s)

	var tKeyList []string = s.Storage.SelectTDkeyList()
	for i := 0; i < len(tKeyList); i++ {
		keyMap[tKeyList[i]] = tKeyList[i]
	}
	log.Info().Msgf("loading keyMap count:'%d'", len(tKeyList))

	loggedRouter := handlers.LoggingHandler(os.Stdout, s.router)

	go s.CleanStackMap()

	secondsSinceEpoch := time.Now().Unix()
	rand.Seed(uint64(secondsSinceEpoch))

	log.Info().Msgf("Starting server: '%s'", s.Listen)
	if err := http.ListenAndServe(s.Listen, handlers.CORS(credentials, methods, origins)(loggedRouter)); err != nil {
		log.Fatal().Msgf("Error starting server: %v", err)
	}

}
