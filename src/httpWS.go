package main

import (
	"net"
	"net/http"

	gsessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

/*
	Validate that they have logged in before opening a WebSocket connection

	Essentially, all we need to do is check to see if they have any cookie
	values stored, because that implies that they got through the "httpLogin"
	less than 5 seconds ago. But we also do a few other checks to be thorough.
*/

func httpWS(c *gin.Context) {
	// Local variables
	w := c.Writer
	r := c.Request

	// Parse the IP address
	var ip string
	if v, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
		log.Error("Failed to parse the IP address in the WebSocket function:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else {
		ip = v
	}

	// Check to see if their IP is banned
	if userIsBanned, err := db.BannedIPs.Check(ip); err != nil {
		log.Error("Failed to check to see if the IP \""+ip+"\" is banned:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if userIsBanned {
		log.Info("IP \"" + ip + "\" tried to establish a WebSocket connection, but they are banned.")
		http.Error(w, "Your IP address has been banned. Please contact an administrator if you think this is a mistake.", http.StatusUnauthorized)
		return
	}

	// If they have logged in, their cookie should have values that we set in httpLogin.go
	session := gsessions.Default(c)
	var userID int
	if v := session.Get("userID"); v == nil {
		log.Error("Unauthorized WebSocket handshake detected from \"" + ip + "\" (failed userID check).")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	} else {
		userID = v.(int)
	}
	var username string
	if v := session.Get("username"); v == nil {
		log.Error("Unauthorized WebSocket handshake detected from \"" + ip + "\" (failed username check).")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	} else {
		username = v.(string)
	}
	var admin int
	if v := session.Get("admin"); v == nil {
		log.Error("Unauthorized WebSocket handshake detected from \"" + ip + "\" (failed admin check).")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	} else {
		admin = v.(int)
	}

	// Check for sessions that belong to orphaned accounts
	if exists, user, err := db.Users.Get(username); err != nil {
		log.Error("Failed to get user \""+username+"\":", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if !exists {
		log.Error("User \"" + username + "\" does not exist in the database; they are trying to establish a WebSocket connection with an orphaned account.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	} else if userID != user.ID {
		log.Error("User \"" + username + "\" exists in the database, but they are trying to establish a WebSocket connection with an account ID that does not match the ID in the database.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// If they got this far, they are a valid user
	// Transfer the values from the login cookie into WebSocket session variables
	keys := make(map[string]interface{})
	keys["userID"] = userID
	keys["username"] = username
	keys["admin"] = admin
	keys["currentGame"] = -1 // By default, the user is not in any games
	keys["status"] = "Lobby" // By default, the user is in the lobby

	// Validation succeeded, so establish the WebSocket connection
	if err := m.HandleRequestWithKeys(w, r, keys); err != nil {
		log.Error("Failed to establish the WebSocket connection for user \""+username+"\":", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
