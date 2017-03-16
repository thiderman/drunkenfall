package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"fmt"
	"github.com/thiderman/drunkenfall/websockets"
	"golang.org/x/net/websocket"
	"strings"
)

const (
	semi   = "semi"
	final  = "final"
	tryout = "tryout"
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

// GeneralRedirect is an explicit permission failure
type GeneralRedirect JSONMessage

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

// PeopleList returns a list with users
type PeopleList struct {
	People []*Person `json:"people"`
}

// NewRequest is the request to make a new tournament
type NewRequest struct {
	Name      string    `json:"name"`
	ID        string    `json:"id"`
	Scheduled time.Time `json:"scheduled"`
	Fake      bool      `json:"fake"`
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

func init() {
	// To get line numbers in log output
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
		t, err = NewTournament(req.Name, req.ID, req.Scheduled, s)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
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
	ps := PersonFromSession(s, r)
	p := NewPlayer(ps)

	err := tm.AddPlayer(p)
	if err != nil {
		PermissionFailure(w, r, err.Error())
		return
	}

	log.Printf("%s has joined %s!", p.Name(), tm.Name)
	s.Redirect(w, tm.URL())
}

// EditHandler shows the tournament view and handles tournaments
func (s *Server) EditHandler(w http.ResponseWriter, r *http.Request) {
	if !HasPermission(r, PermissionProducer) {
		PermissionFailure(w, r, "You need to be very hax to edit a tournament")
		return
	}

	ps := PersonFromSession(s, r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	t, err := LoadTournament([]byte(data), s.DB)
	if err != nil {
		log.Fatal(err)
	}

	err = s.DB.OverwriteTournament(t)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s has edited %s!", ps.Nick, t.ID)
	t.Persist()
	s.Redirect(w, t.URL())
}

// BackfillSemisHandler shows the tournament view and handles tournaments
func (s *Server) BackfillSemisHandler(w http.ResponseWriter, r *http.Request) {
	if !HasPermission(r, PermissionJudge) {
		PermissionFailure(w, r, "You need to be a judge to backfill")
		return
	}

	tm := s.getTournament(r)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		PermissionFailure(w, r, err.Error())
		return
	}

	spl := strings.Split(string(data), ",")
	err = tm.BackfillSemis(spl)

	if err != nil {
		PermissionFailure(w, r, err.Error())
		return
	}

	s.Redirect(w, tm.URL())
}

// ReshuffleHandler reshuffles the player order of the tournament
func (s *Server) ReshuffleHandler(w http.ResponseWriter, r *http.Request) {
	if !HasPermission(r, PermissionProducer) {
		PermissionFailure(w, r, "You need to be a producer to reshuffle")
		return
	}

	tm := s.getTournament(r)
	err := tm.Reshuffle()

	if err != nil {
		PermissionFailure(w, r, err.Error())
		return
	}

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

	tm.Persist() // TODO(thiderman): Move into NextMatch, probably. Should not be here.
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	s.Redirect(w, m.URL())
}

// MatchFunctor is a common function for usage in MatchHandler
type MatchFunctor func(w http.ResponseWriter, r *http.Request, match *Match) error

// MatchHandler is the common function for match operations.
func (s *Server) MatchHandler(w http.ResponseWriter, r *http.Request, functor MatchFunctor) {
	if !HasPermission(r, PermissionJudge) {
		PermissionFailure(w, r, "Cannot modify match unless judge or above")
		return
	}

	m := s.getMatch(r)
	err := functor(w, r, m)
	if err != nil {
		msg := err.Error()
		log.Print(msg)
		ErrorResponse(w, r, msg)
		return
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

// MatchEndHandler ends matches
func (s *Server) MatchEndHandler(w http.ResponseWriter, r *http.Request) {
	s.MatchHandler(w, r, func(w http.ResponseWriter, r *http.Request, m *Match) error {
		if !m.IsStarted() {
			errorMsg := fmt.Sprintf("Cannot end the match `%s` that is in not started.", m.String())
			return errors.New(errorMsg)
		}
		log.Printf("%s ended", m.String())
		err := m.End()
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
}

// MatchStartHandler starts matches
func (s *Server) MatchStartHandler(w http.ResponseWriter, r *http.Request) {
	s.MatchHandler(w, r, func(w http.ResponseWriter, r *http.Request, m *Match) error {
		log.Print("Trying to start match!")
		if m.IsStarted() {
			log.Print("Trying to send error. Wäääääääää")
			errorMsg := fmt.Sprintf("Cannot start the match `%s` that is already in progress.", m.String())
			return errors.New(errorMsg)
		}
		log.Printf("%s started", m.String())
		err := m.Start()
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
}

// MatchResetHandler starts matches
func (s *Server) MatchResetHandler(w http.ResponseWriter, r *http.Request) {
	s.MatchHandler(w, r, func(w http.ResponseWriter, r *http.Request, m *Match) error {
		if !m.IsStarted() {
			errorMsg := fmt.Sprintf("Cannot reset the match `%s` that is not started yet.", m.String())
			return errors.New(errorMsg)
		}

		err := m.Reset()
		if err != nil {
			return err
		}

		log.Printf("%s has been reset", m.String())
		return nil
	})
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

// ClearTournamentHandler removes all test tournaments
func (s *Server) ClearTournamentHandler(w http.ResponseWriter, r *http.Request) {
	err := s.DB.ClearTestTournaments()
	if err != nil {
		log.Fatal(err)
	}
	s.TournamentListHandler(w, r)
}

// ToggleHandler manages who's in the tournament or not
func (s *Server) ToggleHandler(w http.ResponseWriter, r *http.Request) {
	if !HasPermission(r, PermissionJudge) {
		PermissionFailure(w, r, "You need to be a manager to change joins")
		return
	}

	vars := mux.Vars(r)
	id := vars["person"]
	t := s.getTournament(r)

	t.TogglePlayer(id)
}

// SetTimeHandler sets the pause time for the next match
func (s *Server) SetTimeHandler(w http.ResponseWriter, r *http.Request) {
	if !HasPermission(r, PermissionCommentator) {
		PermissionFailure(w, r, "You need to be a commentator to change times")
		return
	}

	vars := mux.Vars(r)
	t := s.getTournament(r)
	x, err := strconv.Atoi(vars["time"])
	if err != nil {
		log.Fatal(err)
	}

	m, err := t.NextMatch()
	if err != nil {
		log.Fatal(err)
	}

	// If the match is already started, we need to bail
	if m.IsScheduled() {
		PermissionFailure(w, r, "Match already started")
		return
	}

	m.SetTime(x)
	s.Redirect(w, m.URL())
}

// PeopleHandler returns a list of all the players registered in the app
func (s *Server) PeopleHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.DB.LoadPeople(); err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(&PeopleList{
		People: s.DB.People,
	})
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// BuildRouter sets up the routes
func (s *Server) BuildRouter(ws *websockets.Server) http.Handler {
	n := mux.NewRouter()
	a := n.PathPrefix("/api").Subrouter()
	r := a.PathPrefix("/towerfall").Subrouter()

	r.HandleFunc("/people/", s.PeopleHandler)
	r.HandleFunc("/tournament/", s.TournamentListHandler)
	r.HandleFunc("/tournament/clear/", s.ClearTournamentHandler)
	r.HandleFunc("/tournament/{id}/", s.TournamentHandler)
	// TODO(thiderman): Normalize for all to use /tournament
	r.HandleFunc("/new/", s.NewHandler)
	r.HandleFunc("/{id}/start/", s.StartTournamentHandler)
	r.HandleFunc("/{id}/usurp/", s.UsurpTournamentHandler)
	r.HandleFunc("/{id}/join/", s.JoinHandler)
	r.HandleFunc("/{id}/edit/", s.EditHandler)
	r.HandleFunc("/{id}/reshuffle/", s.ReshuffleHandler)
	r.HandleFunc("/{id}/backfill/", s.BackfillSemisHandler)
	r.HandleFunc("/{id}/toggle/{person}", s.ToggleHandler)
	r.HandleFunc("/{id}/time/{time}", s.SetTimeHandler)
	r.HandleFunc("/{id}/next/", s.NextHandler)

	// Install the websockets
	r.Handle("/auto-updater", websocket.Handler(ws.OnConnected))

	// Handle Facebook
	s.FacebookRouter(a)

	m := r.PathPrefix("/tournament/{id}/{kind:(?:tryout|runnerup|semi|final)}/{index:[0-9]+}").Subrouter()

	m.HandleFunc("/end/", s.MatchEndHandler)
	m.HandleFunc("/start/", s.MatchStartHandler)
	m.HandleFunc("/reset/", s.MatchResetHandler)

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
	index, err := strconv.Atoi(vars["index"])
	if err != nil {
		log.Printf("Translation went horribly wrong from %s. Index is now 0. Error is: %s", vars["index"], err)
		log.Printf("Vars are: %s", vars)
	}
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
	GeneralResponse(w, r, http.StatusUnauthorized, msg)
}

// ErrorResponse returns an error with the statuscode of 400
func ErrorResponse(w http.ResponseWriter, r *http.Request, msg string) {
	GeneralResponse(w, r, http.StatusBadRequest, msg)
}

// GeneralResponse returns an error with the statuscode of status, status being
// the input of the function. Also redirects the user to the best of its ability
// to / (meaning errors are completely ignored :') ).
func GeneralResponse(w http.ResponseWriter, r *http.Request, status int, msg string) {
	data, err := json.Marshal(GeneralRedirect{
		Message:  msg,
		Redirect: "/",
	})
	if err != nil {
		log.Print(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func main() {
	// Instantiate the database
	db, err := NewDatabase("production.db")
	if err != nil {
		log.Fatal(err)
	}

	// Apply any applicable migrations
	err = Migrate(db.DB)
	if err != nil {
		log.Fatal(err)
	}

	// Create the server instance...
	s := NewServer(db)
	// ...and give the db a reference to the server.
	// Not the cleanest, but y'know... here we are.
	db.Server = s

	// Load tournaments from the database
	err = db.LoadTournaments()
	if err != nil {
		log.Fatal(err)
	}

	// Set up the paths and the websocket listeners
	s.RegisterHandlersAndListeners()

	// Actually start serving
	if err := s.Serve(); err != nil {
		log.Fatal(err)
	}
}
