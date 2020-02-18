// userAdministration
/*
 * Diese Klasse händelt die Nutzeranfragen zur Registrierung, Login, Logout
 * sowie Nutzer löschen. Für die Auhentifizierung wird das Paket Authenticate
 * verwendet, welches modular in Übung06 erstellt wurde.
 */
package main

import (
	"encoding/json"
	"net/http"
	auth "picx/packages/authentificate"
	"strings"

	"github.com/globalsign/mgo/bson"
)

//type um die id des Users auszulesen (für den Cookie)
type readUserTy struct {
	Id       bson.ObjectId `bson:"_id"`
	Name     string        `bson:"name"`
	Password string
}

/*------------------------------------------------------------------------------
					Authentifizierung des Nutzers
-------------------------------------------------------------------------------*/
func authUserHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	_, err := r.Cookie(cookieName)

	//wenn cookie noch nicht gesetzt ist fortfahren
	if err != nil {
		//request-Daten lesen:
		username := strings.ToLower(r.FormValue("name"))
		password := r.FormValue("password")
		method := r.FormValue("method")

		//alle Nutzer aus der Datenbank auslesen
		allUser := []auth.UserTy{}
		colUser := dbSession.DB(dbName).C(dbColUser)
		colUser.Find(nil).All(&allUser)

		//User registrieren mit auth paket
		if method == "register" {
			validRegistration, msg := auth.RegisterUser(username, password, allUser)

			//user in die Datenbank eintragen
			if validRegistration {
				newUser := auth.UserTy{
					Name:     username,
					Password: password,
				}
				colUser.Insert(newUser)
				msg = "User " + username + " erfolgreich registriert!"
				sendMessageToClient(w, msg)
			} else {
				sendMessageToClient(w, msg)
			}

			//Login des Users mit auth paket
		} else if method == "login" {

			user, msg := auth.LoginUser(username, password, allUser)
			//login erfolgreich
			if user.Name != "" {
				//userID auslesen
				userWithId := readUserTy{}
				colUser.Find(bson.M{"name": bson.M{"$eq": user.Name}}).One(&userWithId)

				//cookie wird mit userID gesetzt
				newCookie := http.Cookie{
					Name:  cookieName,
					Value: userWithId.Id.Hex(),
				}
				http.SetCookie(w, &newCookie)

				//Username des eingelogten Nutzers wird an den Client gesendet
				userInfo := messageTy{
					Username: username,
				}
				//format to json and send it
				jData, _ := json.Marshal(userInfo)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jData)

			} else {
				sendMessageToClient(w, msg)
			}
		}
		//Cookie bereits gesetzt
	} else {
		sendMessageToClient(w, "Nur ein Login pro Browser möglich!")
	}
}

/*------------------------------------------------------------------------------
							Logout des Nutzers
-------------------------------------------------------------------------------*/
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	//cookie wird gelöscht
	newCookie := http.Cookie{
		Name:   cookieName,
		MaxAge: -1,
	}
	http.SetCookie(w, &newCookie)
}

/*------------------------------------------------------------------------------
			Alle Daten des Nutzers werden aus der DB entfernt
-------------------------------------------------------------------------------*/
func deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	//Collections der Sammlungen und Bildinformationen löschen
	dbSession.DB(dbName).C(dbColInfo + cookie.Value).DropCollection()
	dbSession.DB(dbName).C(dbColCollections + cookie.Value).DropCollection()
	dbSession.DB(dbName).C(dbColPools + cookie.Value).DropCollection()

	//Alle Bilder aus dem Gridfs des Nutzers löschen
	gridfsImages := dbSession.DB(dbName).GridFS(dbGridFsBilder + cookie.Value)
	images := []infoTy{}
	gridfsImages.Find(nil).All(&images)
	for i := 0; i < len(images); i++ {
		gridfsImages.RemoveId(images[i].Id)
	}

	//Alle Thumbnails aus dem Gridfs des Nutzers löschen
	gridfsThumbs := dbSession.DB(dbName).GridFS(dbGridFsThumbs + cookie.Value)
	thumbs := []infoTy{}
	gridfsThumbs.Find(nil).All(&thumbs)
	for i := 0; i < len(thumbs); i++ {
		gridfsThumbs.RemoveId(thumbs[i].Id)
	}

	//Nutzer aus der Datenbank entfernen
	userCol := dbSession.DB(dbName).C(dbColUser)
	userCol.RemoveId(bson.ObjectIdHex(cookie.Value))

	sendMessageToClient(w, "Account erfolgreich gelöscht.")
	logoutHandler(w, r) //user wird ausgelogt
}
