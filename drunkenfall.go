package main

import (
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"fmt"
	"github.com/thiderman/drunkenfall/websockets"
	"golang.org/x/net/websocket"
)

// Setup variables for the cookies. Can be used outside of this file.
var (
	CookieStoreKey = []byte("dtf")
	CookieStore    = sessions.NewCookieStore(CookieStoreKey)
)

// Server is an abstraction that runs via a web interface
type Server struct {
	DB     *Database
	router http.Handler
	logger http.Handler
	ws     *websockets.Server
}

// JSONMessage defines a message to be returned to the frontend
type JSONMessage struct {
	Message  string `json:"message"`
	Redirect string `json:"redirect"`
}

// PermissionRedirect is an explicit permission failure
type PermissionRedirect JSONMessage

// TournamentMessage returns a single tournament
type TournamentMessage struct {
	Tournament *Tournament `json:"tournament"`
}

// UpdateMessage returns an update to the current tournament
type UpdateMessage TournamentMessage

// UpdateMatchMessage returns an update to the current match
type UpdateMatchMessage struct {
	Match *Match `json:"match"`
}

// TournamentList returns a list with tournaments
type TournamentList struct {
	Tournaments []*Tournament `json:"tournaments"`
}

// UpdateStateMessage returns an update to the current match
type UpdateStateMessage TournamentList

// NewRequest is the request to make a new tournament
type NewRequest struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Fake bool   `json:"fake"`
}

// CommitPlayer is one state for a player in a commit message
type CommitPlayer struct {
	Ups    int    `json:"ups"`
	Downs  int    `json:"downs"`
	Shot   bool   `json:"shot"`
	Reason string `json:"reason"`
}

// CommitRequest is a request to commit a match state
type CommitRequest struct {
	State []CommitPlayer `json:"state"`
}

// NewServer instantiates a server with an active database
func NewServer(db *Database) *Server {
	s := Server{DB: db}
	s.ws = websockets.NewServer()
	s.router = s.BuildRouter(s.ws)

	return &s
}

// RegisterHandlersAndListeners registers the routes and the websocket listeners
func (s *Server) RegisterHandlersAndListeners() {
	http.Handle("/", s.router)
	s.logger = handlers.LoggingHandler(os.Stdout, s.router)

	// Also websocket listener
	go s.ws.Listen()
}

// NewHandler shows the page to create a new tournament
func (s *Server) NewHandler(w http.ResponseWriter, r *http.Request) {
	var req NewRequest
	var t *Tournament

	if !HasPermission(r, PermissionProducer) {
		PermissionFailure(w, r, "Cannot create match unless producer")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Fatal(err)
	}

	if req.Fake {
		t = SetupFakeTournament(s)
	} else {
		t, _ = NewTournament(req.Name, req.ID, s)
	}

	log.Printf("Created tournament %s!", t.Name)

	s.DB.Tournaments = append(s.DB.Tournaments, t)
	s.DB.tournamentRef[t.ID] = t

	s.Redirect(w, t.URL())
}

// TournamentHandler returns the current state of the tournament
func (s *Server) TournamentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	tm := s.DB.tournamentRef[vars["id"]]
	out := &TournamentMessage{tm}

	data, err := json.Marshal(out)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// JoinHandler shows the tournament view and handles tournaments
func (s *Server) JoinHandler(w http.ResponseWriter, r *http.Request) {
	if !HasPermission(r, PermissionPlayer) {
		PermissionFailure(w, r, "You need to sign in to join a tournament")
		return
	}

	tm := s.getTournament(r)
	p := PersonFromSession(s, r)

	err := tm.AddPlayer(p)
	if err != nil {
		PermissionFailure(w, r, err.Error())
		return
	}

	log.Printf("%s has joined %s!", p.Name, tm.Name)
	s.Redirect(w, tm.URL())
}

// StartTournamentHandler starts tournaments
func (s *Server) StartTournamentHandler(w http.ResponseWriter, r *http.Request) {
	if !HasPermission(r, PermissionCommentator) {
		PermissionFailure(w, r, "Cannot start tournament unless commentator or above")
		return
	}

	tm := s.getTournament(r)
	err := tm.StartTournament()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	s.Redirect(w, tm.URL())
}

// UsurpTournamentHandler usurps tournaments
func (s *Server) UsurpTournamentHandler(w http.ResponseWriter, r *http.Request) {
	if !HasPermission(r, PermissionCommentator) {
		PermissionFailure(w, r, "Cannot usurp tournament unless commentator or above")
		return
	}

	tm := s.getTournament(r)
	err := tm.UsurpTournament()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	s.Redirect(w, tm.URL())
}

// NextHandler sets the tournament up to play the next match
func (s *Server) NextHandler(w http.ResponseWriter, r *http.Request) {
	if !HasPermission(r, PermissionCommentator) {
		PermissionFailure(w, r, "Cannot goto next match unless commentator or above")
		return
	}

	tm := s.getTournament(r)
	m, err := tm.NextMatch()
	tm.SetCurrent(m)

	tm.Persist() // TODO(thiderman): Move into NextMatch, probably. Should not be here.
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	s.Redirect(w, m.URL())
}

// MatchToggleHandler starts and stops matches
func (s *Server) MatchToggleHandler(w http.ResponseWriter, r *http.Request) {
	// TODO(thiderman): This should really be two different methods.
	if !HasPermission(r, PermissionJudge) {
		PermissionFailure(w, r, "Cannot start match unless judge or above")
		return
	}

	m := s.getMatch(r)
	if !m.IsStarted() {
		log.Printf("%s started", m.String())
		m.Start()
	} else {
		log.Printf("%s ended", m.String())
		m.End()
	}

	data, err := json.Marshal(UpdateMessage{
		Tournament: m.Tournament,
	})
	if err != nil {
		log.Print(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// MatchCommitHandler commits a single round of a match
func (s *Server) MatchCommitHandler(w http.ResponseWriter, r *http.Request) {
	if !HasPermission(r, PermissionJudge) {
		PermissionFailure(w, r, "Cannot commit match unless judge or above")
		return
	}

	var req CommitRequest

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Print(err)
		return
	}

	c := NewMatchCommit(req)
	m := s.getMatch(r)
	m.Commit(c)

	data, err := json.Marshal(UpdateMatchMessage{
		Match: m,
	})
	if err != nil {
		log.Print(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)

	return
}

// TournamentListHandler returns a list of all tournaments
func (s *Server) TournamentListHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(&TournamentList{
		Tournaments: s.DB.Tournaments,
	})
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// BuildRouter sets up the routes
func (s *Server) BuildRouter(ws *websockets.Server) http.Handler {
	n := mux.NewRouter()
	a := n.PathPrefix("/api").Subrouter()
	r := a.PathPrefix("/towerfall").Subrouter()

	r.HandleFunc("/tournament/", s.TournamentListHandler)
	r.HandleFunc("/tournament/{id}/", s.TournamentHandler)
	// TODO: Normalize for all to use /tournament
	r.HandleFunc("/new/", s.NewHandler)
	r.HandleFunc("/{id}/start/", s.StartTournamentHandler)
	r.HandleFunc("/{id}/usurp/", s.UsurpTournamentHandler)
	r.HandleFunc("/{id}/join/", s.JoinHandler)
	r.HandleFunc("/{id}/next/", s.NextHandler)

	// Install the websockets
	r.Handle("/auto-updater", websocket.Handler(ws.OnConnected))

	// Handle Facebook
	s.FacebookRouter(a)

	m := r.PathPrefix("/tournament/{id}/{kind:(tryout|runnerup|semi|final)}/{index:[0-9]+}").Subrouter()
	m.HandleFunc("/toggle/", s.MatchToggleHandler)
	m.HandleFunc("/commit/", s.MatchCommitHandler)

	return n
}

// Serve serves forever
func (s *Server) Serve() error {
	log.Print("Listening on :42001")
	return http.ListenAndServe(":42001", s.logger)
}

// SendWebsocketUpdate sends an update to all listening sockets
func (s *Server) SendWebsocketUpdate() {
	msg := websockets.Message{
		Data: UpdateStateMessage{
			Tournaments: s.DB.Tournaments,
		},
	}

	s.ws.SendAll(&msg)
}

func (s *Server) getMatch(r *http.Request) *Match {
	var m *Match
	vars := mux.Vars(r)

	tm := s.DB.tournamentRef[vars["id"]]
	kind := vars["kind"]
	index, _ := strconv.Atoi(vars["index"])

	if kind == "tryout" {
		m = tm.Tryouts[index]
	} else if kind == "semi" {
		m = tm.Semis[index]
	} else if kind == "final" {
		m = tm.Final
	}

	return m
}

func (s *Server) getTournament(r *http.Request) *Tournament {
	vars := mux.Vars(r)
	tm := s.DB.tournamentRef[vars["id"]]
	return tm
}

// Redirect creates a JSON Redirect
func (s *Server) Redirect(w http.ResponseWriter, url string) {
	data, err := json.Marshal(JSONMessage{
		Redirect: url,
	})
	if err != nil {
		log.Print(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// HasPermission checks that the user is allowed to do an action
func HasPermission(r *http.Request, lvl int) bool {
	s, _ := CookieStore.Get(r, "session")
	l, ok := s.Values["userlevel"]
	if !ok {
		// log.Print("Userlevel missing for auth")
		return false
	}

	log.Print(fmt.Sprintf("Auth check: %s: %d", s.Values, lvl))
	return l.(int) >= lvl
}

// PermissionFailure returns an error 401
func PermissionFailure(w http.ResponseWriter, r *http.Request, msg string) {
	data, err := json.Marshal(PermissionRedirect{
		Message:  msg,
		Redirect: "/",
	})
	if err != nil {
		log.Print(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write(data)
}

func main() {
	db, err := NewDatabase("production.db")
	if err != nil {
		log.Fatal(err)
	}

	s := NewServer(db)
	db.Server = s

	err = db.LoadTournaments()
	if err != nil {
		log.Fatal(err)
	}

	// Set up the paths and the websocket listeners
	s.RegisterHandlersAndListeners()
	err = s.Serve()

	if err != nil {
		log.Fatal(err)
	}
}
