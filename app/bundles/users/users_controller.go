package users

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krishamoud/game/app/common/controller"
	"github.com/krishamoud/game/app/common/middleware"
)

// Controller struct
type Controller struct {
	common.Controller
}

// Index func return all users
func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
	var err error
	result, err := indexUsers()
	if c.CheckError(err, http.StatusInternalServerError, w) {
		return
	}
	c.SendJSON(
		w,
		r,
		result,
		http.StatusOK,
	)
}

// New shows the new user page
func (c *Controller) New(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Origin, X-Requested-With, Content-Type, Accept")
	c.SendJSON(
		w,
		r,
		[]int{0, 1, 2},
		http.StatusOK,
	)
}

// Create saves a new user to the database
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := createUser(email, password)
	if c.CheckError(err, http.StatusBadRequest, w) {
		return
	}
	c.SendJSON(
		w,
		r,
		user,
		http.StatusOK,
	)
}

// Show a single user
func (c *Controller) Show(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	userID := vars["userId"]
	ctx := r.Context()
	fmt.Println(ctx)
	reqUser := middleware.UserID(ctx)
	result, err := showUser(userID, reqUser)

	if c.CheckError(err, http.StatusNotFound, w) {
		return
	}
	c.SendJSON(
		w,
		r,
		result,
		http.StatusOK,
	)
}

// Edit page
func (c *Controller) Edit(w http.ResponseWriter, r *http.Request) {}

// Update user doc and save to database
func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {}

// Destroy a user
func (c *Controller) Destroy(w http.ResponseWriter, r *http.Request) {}

// Auth a user
func (c *Controller) Auth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := authenticateUser(email, password)
	if c.CheckError(err, http.StatusUnauthorized, w) {
		return
	}
	c.SendJSON(
		w,
		r,
		user,
		http.StatusOK,
	)
}
