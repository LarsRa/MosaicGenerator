// picx
/*
 * Die Mainanwendung initialisiert die einzelnen Handler für die AJAX-Anfragen.
 * Die Datenbankvariablen für die Namenskonvention werden hier festgelegt und
 * initial die Datenbankverbindung aufgebaut. Es wird eine Funktion zum Error-
 * handling und zum senden einer Nachicht im JSON-Format an den Clienten bereit-
 * gestellt.
 */
package main

import (
	"encoding/json"
	"fmt"
	"html/template"

	"net/http"

	"github.com/globalsign/mgo"
)

//
// template datei übergeben
var t = template.Must(template.ParseGlob("tmpl/index.html"))

//variablen der datenbank
var dbSession *mgo.Session
var dbName = "HA19_lars_raschke_591098"
var dbColUser = "USER"
var dbGridFsBilder = "IMAGES"
var dbGridFsThumbs = "THUMBNAILS"
var dbColPools = "POOLS"
var dbColCollections = "COLLECTIONS"
var dbColInfo = "IMAGEINFOS"

var cookieName = "picxCookie"

var err error

//Um dem Clienten Rückmeldungen zu Aktionen zu senden
type messageTy struct {
	Text     string
	Username string
}

func main() {
	//Handler für Nutzerverwaltung werden registriert
	http.HandleFunc("/picx", rootHandler)
	http.HandleFunc("/authentificate", authUserHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/deleteAccount", deleteAccountHandler)

	//Handler für Bilderverwaltung werden registriert
	http.HandleFunc("/uploadImages", uploadImagesHandler)
	http.HandleFunc("/getImages", getImagesHandler)
	http.HandleFunc("/getImage", getImageHandler)
	http.HandleFunc("/resizeImages", resizeImagesHandler)
	http.HandleFunc("/getImageInfo", getImageInfoHandler)
	http.HandleFunc("/removeImages", removeImagesHandler)
	http.HandleFunc("/downloadMosaic", downloadMosaicHandler)

	//Handler für Gruppenverwaltung werden registriert
	http.HandleFunc("/createGroup", createNewGroupHandler)
	http.HandleFunc("/getAllGroups", getAllGroupsHandler)
	http.HandleFunc("/addToGroup", addToGroupHandler)
	http.HandleFunc("/getPoolInfo", getPoolInfoHandler)
	http.HandleFunc("/removeGroup", removeGroupHandler)

	//Handler für Generatoren werden registriert
	http.HandleFunc("/generateImages", generateImagesHandler)
	http.HandleFunc("/generateMosaic", generateMosaicHandler)

	//File Server
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":4242", nil)
}

/*------------------------------------------------------------------------------
				Verbindung zur Datenbank initial aufbauen
-------------------------------------------------------------------------------*/
func init() {
	var err error
	//Verbindung zum Mongo-DBMS:
	//dbSession, err := mgo.Dial("mongodb://borsti.inf.fh-flensburg.de:27017")
	dbSession, err = mgo.Dial("localhost:27017")
	if err != nil {
		fmt.Println("error")
	}
}

/*------------------------------------------------------------------------------
						Indexseite wird geladen
-------------------------------------------------------------------------------*/
func rootHandler(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "index.html", nil) // Seite senden
}

//Funktion zum Errorhandling aus dem Vorlesungsbeispiel
func check_ResponseToHTTP(err error, w http.ResponseWriter) {
	if err != nil {
		fmt.Fprintln(w, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

/*------------------------------------------------------------------------------
					Nachricht wird an den Client geschickt
-------------------------------------------------------------------------------*/
func sendMessageToClient(w http.ResponseWriter, msg string) {

	clientMessage := messageTy{
		Text: msg,
	}
	//format to json
	jData, _ := json.Marshal(clientMessage)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}
