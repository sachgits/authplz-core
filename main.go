package main

import "os"

//import "strings"
import "fmt"
import "log"
import "net/http"

import "encoding/gob"

import "github.com/gocraft/web"

//import "github.com/kataras/iris"
import "github.com/gorilla/sessions"
import "github.com/gorilla/context"

import "github.com/ryankurte/authplz/usercontroller"
import "github.com/ryankurte/authplz/token"
import "github.com/ryankurte/authplz/datastore"

// Application global context
// TODO: this could be split and bound by module
type AuthPlzGlobalCtx struct {
	port            string
	address         string
	userController  *usercontroller.UserController
	tokenController *token.TokenController
	sessionStore    *sessions.CookieStore
}

// Application handler context
type AuthPlzCtx struct {
	global  *AuthPlzGlobalCtx
	session *sessions.Session
	userid  string
}

// Convenience type to describe middleware functions
type MiddlewareFunc func(ctx *AuthPlzCtx, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc)

// Bind global context object into the router context
func BindContext(globalCtx *AuthPlzGlobalCtx) MiddlewareFunc {
	return func(ctx *AuthPlzCtx, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
		ctx.global = globalCtx
		next(rw, req)
	}
}

// User session layer
func (ctx *AuthPlzCtx) SessionMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	session, err := ctx.global.sessionStore.Get(req.Request, "user-session")
	if err != nil {
		next(rw, req)
		return
	}

	// Save session for further use
	ctx.session = session

	// Load user from session if set
	// TODO: this will be replaced with sessions when implemented
	if session.Values["userId"] != nil {
		fmt.Println("userId found")
		//TODO: find user account
		ctx.userid = session.Values["userId"].(string)
	}

	//session.Save(r, w)
	next(rw, req)
}

func (c *AuthPlzCtx) RequireAccountMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	if c.userid == "" {
		c.WriteApiResult(rw, ApiResultError, "You must be signed in to view this page")
	} else {
		next(rw, req)
	}
}

func (c *AuthPlzCtx) LoginUser(u *datastore.User, rw web.ResponseWriter, req *web.Request) {
	c.session.Values["userId"] = u.UUID
	c.session.Save(req.Request, rw)
}

func (c *AuthPlzCtx) LogoutUser(rw web.ResponseWriter, req *web.Request) {
	c.session.Options.MaxAge = -1
	c.session.Save(req.Request, rw)
}

type AuthPlzServer struct {
	address string
	port    string
	ds      *datastore.DataStore
	ctx     AuthPlzGlobalCtx
	router  *web.Router
}

func NewServer(address string, port string, db string) *AuthPlzServer {
	server := AuthPlzServer{}

	server.address = address
	server.port = port

	gob.Register(&token.TokenClaims{})

	// Attempt database connection
	ds, err := datastore.NewDataStore(db)
	if err != nil {
		panic("Error opening database")
	}
	server.ds = ds;

	// Create session store
	sessionStore := sessions.NewCookieStore([]byte("something-very-secret"))

	// Create controllers
	uc := usercontroller.NewUserController(server.ds, nil)
	tc := token.NewTokenController(server.address, "something-also-secret")

	// Create a global context object
	server.ctx = AuthPlzGlobalCtx{port, address, &uc, &tc, sessionStore}

	// Create router
	server.router = web.New(AuthPlzCtx{}).
		Middleware(BindContext(&server.ctx)).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Middleware((*AuthPlzCtx).SessionMiddleware).
		Post("/api/login", (*AuthPlzCtx).Login).
		Post("/api/create", (*AuthPlzCtx).Create).
		Post("/api/action", (*AuthPlzCtx).Action).
		Get("/api/action", (*AuthPlzCtx).Action).
		Get("/api/logout", (*AuthPlzCtx).Logout).
		Get("/api/status", (*AuthPlzCtx).Status).
		Get("/api/test", (*AuthPlzCtx).Test)

	return &server
}

func (server *AuthPlzServer) Start() {
	// Start listening
	fmt.Println("Listening at: " + server.port)
	log.Fatal(http.ListenAndServe(server.address+":"+server.port, context.ClearHandler(server.router)))
}

func (server *AuthPlzServer) Close() {
	server.ds.Close()
}

func main() {
	var port string = "9000"
	var address string = "localhost"
	var dbString string = "host=localhost user=postgres dbname=postgres sslmode=disable password=postgres"

	// Parse environmental variables
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	server := NewServer(address, port, dbString)

	server.Start()
}
