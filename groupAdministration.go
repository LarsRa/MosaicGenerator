// groupAdministration
/*
 * Diese Klasse bearbeitet alle Anfragen des Clients in hinsicht auf die
 * Verwaltung von Gruppen (Pools und Sammlungen). Es können Gruppen erstellt,
 * gelöscht, Bilder hinzugefügt, Informationen ausgelesen und die mittlere
 * Farbverteilung eines Pools ausgelesen werden.
 */
package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/globalsign/mgo/bson"
)

//Struct um Gruppen(Pools und Sammlungen) zu erstellen
type groupTy struct {
	Size        int      `bson:"size"`
	Name        string   `bson:"name"`
	ColorValues []int    `bson:"colorvalues"` //Reihenfolge R,G,B
	Params      []string `bson:"params"`      //Für generierte Pools(Rechteck,
} //Rechteckgröße)

//Struct um alle Gruppen an den Client zu senden
type sendGroupTy struct {
	Pools       []groupTy
	Collections []groupTy
}

//Struct um gesendete Bilder und Gruppe aus einer Request auszulesen um diese
//dann der passenden Gruppe hinzufügen zu können.
type groupRequest struct {
	ImageIDs   []string
	Group      string
	Collection string
}

/*
 * Der Handler liest übergebene Formdaten aus und ruft die createNewGroup-
 * Funktion auf, welche eine neue Gruppe(Pool oder Sammlung) erstellt.
 */
func createNewGroupHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	category := r.FormValue("category")
	tilesize := r.FormValue("size")
	createNewGroup(w, category, name, tilesize, r, nil)
}

/*
 * Die Funktion erstellt mit den übergebenen Parametern eine Gruppe.
 * Es wird geprüft, ob eine Sammlung oder ein Pool erstellt werden soll und ob
 * der Name dafür bereits verwendet wird. Falls nicht, wird eine neue Gruppe
 * erstellt. Pools oder Sammlungen mit den gleichen Namen sind somit nicht möglich.
 */
func createNewGroup(w http.ResponseWriter, category,
	name, size string, r *http.Request, params []string) bool {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	tilesize, _ := strconv.Atoi(size)
	allGroups := []groupTy{}

	//Prüfen ob ein Pool oder eine Sammlung erstellt werden soll
	colPools := dbSession.DB(dbName).C(dbColPools + cookie.Value)
	colCollections := dbSession.DB(dbName).C(dbColCollections + cookie.Value)
	if category == "pool" {
		//alle bestehenden Pools werden ausgelesen
		colPools.Find(nil).All(&allGroups)
	} else {
		colCollections.Find(nil).All(&allGroups)
	}
	//Prüfen, ob eine Gruppe mit dem Namen bereits existiert
	valid := true
	msg := ""
	for i := 0; i < len(allGroups); i++ {
		if allGroups[i].Name == name {
			valid = false
			msg = "Der Name existiert bereits!"
		}
	}
	//Gruppe wird eingefügt
	if valid {
		newGroup := groupTy{
			Name: name,
			Size: tilesize,
		}
		if category == "pool" {
			colorValues := []int{0, 0, 0}
			newGroup.ColorValues = colorValues
			if params != nil {
				newGroup.Params = params
			}
			colPools.Insert(newGroup)
		} else {
			colCollections.Insert(newGroup)
		}
		msg = "Die Gruppe wurde erfolgreich erstellt!"
	}
	//Rückmeldung über Erfolg an den Client senden
	sendMessageToClient(w, msg)
	return valid
}

/*
 * Der Handler gibt den Client alle bestehenden Gruppen(Pools und Sammlungen) des
 * Nutzers zurück. Diese werden benötigt, um den Selektoren zur Auswahl hinzu-
 * gefügt werden zu können.
 */
func getAllGroupsHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	//alle bestehenden Gruppen werden ausgelesen
	colPools := dbSession.DB(dbName).C(dbColPools + cookie.Value)
	colCollections := dbSession.DB(dbName).C(dbColCollections + cookie.Value)
	allPools := []groupTy{}
	allCollections := []groupTy{}
	colPools.Find(nil).All(&allPools)             //pools werden ausgelesen
	colCollections.Find(nil).All(&allCollections) //sammlungen werden asugelesen

	//Format zum senden an den Client erstellen
	sendGroups := sendGroupTy{
		Pools:       allPools,
		Collections: allCollections,
	}

	//format to json and send it
	jData, _ := json.Marshal(sendGroups)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

/*
 * Der Handler fügt per JSON übergebene Bilder der gefragten Gruppe hinzu.
 * Wenn es sich bei der Gruppe um einen Pool handelt, wird das Bild vorher
 * auf die korrekte Größe skalliert. Die Gruppenzugehörigkeit wird in der
 * Bilderinfo in der DB gespeichert.
 */
func addToGroupHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	//Daten aus JSON Request decodieren
	data := groupRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&data)

	//Prüfen, ob es sich um einen Pool handelt
	pool := groupTy{}

	if data.Collection == "pool" {
		//Größe des Pools auslesen
		colPools := dbSession.DB(dbName).C(dbColPools + cookie.Value)
		colPools.Find(bson.M{"name": bson.M{"$eq": data.Group}}).One(&pool)

		//Bilder aus der DB auf die geforderte Größe skallieren und hochladen
		resizeImagesOfDB(data.ImageIDs, pool.Size, pool.Size, data.Group, r)

		//Rückmeldung über Erfolg an den Client senden
		sendMessageToClient(w, "Erfolgreich skalliert und Pool hinzugefügt!")

	} else { //Bild wird einer Sammlung hinzugefügt
		colInfo := dbSession.DB(dbName).C(dbColInfo + cookie.Value)
		info := infoTy{}
		newInfo := infoTy{}

		//jeder Imageinfo wird der Name der Gruppe hinzugefügt
		for i := 0; i < len(data.ImageIDs); i++ {
			objectId := bson.ObjectIdHex(data.ImageIDs[i])
			colInfo.Find(bson.M{"_id": bson.M{"$eq": objectId}}).One(&info)
			newInfo = addImageToGroup(info, data.Group)
			colInfo.Update(info, newInfo)
		}
		//Rückmeldung über Erfolg an den Client senden
		sendMessageToClient(w, "Erfolgreich der Sammlung hinzugefügt!")
	}
}

/*
 * Fügt einem Bild eine Gruppe zu, wenn das Bild noch nicht in der Gruppe ist.
 * Das selbe Bild mehrmals in einer Gruppe ist somit nicht möglich.
 * Dies geschiet, indem dem Gruppenarray der Bildinformation in der DB der
 * passende Gruppenname hinzugefügt wird.
 */
func addImageToGroup(imageInfo infoTy, group string) infoTy {
	//Gruppennamen der Info hinzufügen
	if group != "all" {
		alredyInGroup := false
		for i := 0; i < len(imageInfo.Groups); i++ {
			if imageInfo.Groups[i] == group {
				alredyInGroup = true
				break
			}
		}

		if !(alredyInGroup) {
			newGroups := append(imageInfo.Groups, group)
			imageInfo.Groups = newGroups
		}
	}
	return imageInfo
}

/*
 * Der Handler liest Informationen zum erfragten Pool aus. Dazu werden die
 * Poolinformationen aus der DB gelesen und die mittleren Farbwerte neu berechent.
 */
func getPoolInfoHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	name := r.FormValue("name")

	//Pool wird ausgelesen
	pool := groupTy{}
	colPool := dbSession.DB(dbName).C(dbColPools + cookie.Value)
	colPool.Find(bson.M{"name": bson.M{"$eq": name}}).One(&pool)

	//Farbwerte des Pools werden berechnet
	if pool.Name != "" {
		calculateAveragePoolColors(name, r)
	}

	//format to json and send it
	jData, _ := json.Marshal(pool)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

/*
 * Der Handler löscht die gefragte Gruppe der Nutzers aus der DB.
 * Dazu wird die Information zum Pool oder Sammlung gelöscht und von jeder
 * Bildinformation der Bilder in dieser Gruppe der Verweis zur Gruppe entfernt.
 * Die Bilder bleiben vorhanden, nur die Gruppe wird entfernt.
 */
func removeGroupHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	groupTyp := r.FormValue("groupTyp")
	groupName := r.FormValue("groupName")

	imagesOfGroup := []infoTy{} //Alle Bilder der Gruppe werden ausgelesen
	colImageInfo := dbSession.DB(dbName).C(dbColInfo + cookie.Value)
	colImageInfo.Find(bson.M{"groups": groupName}).All(&imagesOfGroup)

	//Für jedes Bild der Gruppe den Verweis im Gruppenarray löschen
	for _, imageInfo := range imagesOfGroup {
		for i, group := range imageInfo.Groups {
			if group == groupName {
				updatedGroups := append(imageInfo.Groups[:i], imageInfo.Groups[i+1:]...)
				updatedImageInfo := imageInfo
				updatedImageInfo.Groups = updatedGroups
				colImageInfo.Update(imageInfo, updatedImageInfo)
			}
		}
	}
	//Gruppe aus der Datenbank löschen
	if groupTyp == "Pools" {
		colPools := dbSession.DB(dbName).C(dbColPools + cookie.Value)
		colPools.Remove(bson.M{"name": bson.M{"$eq": groupName}})
	} else {
		colCollections := dbSession.DB(dbName).C(dbColCollections + cookie.Value)
		colCollections.Remove(bson.M{"name": bson.M{"$eq": groupName}})
	}
}

/*
 * Die Funktion berechnet die durchschnittliche Farbverteilung eines Pools.
 * Hierzu werden alle Bilder des Pools ausgelesen, und deren durchschnittliche
 * Farbwerte aufaddiert und daraus ein Durchschnitt gebildet.
 * Die Funktion wird aufgerufen, wenn nach der Poolinformation gefragt wird.
 * (getPoolInfoHandler)
 */
func calculateAveragePoolColors(poolName string, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	//Alle Infos zu Bildern des Pools auslesen
	imagesInfo := []infoTy{}
	colImages := dbSession.DB(dbName).C(dbColInfo + cookie.Value)
	colImages.Find(bson.M{"groups": poolName}).All(&imagesInfo)

	//durcschnittliche Poolfarbe berechnen
	colorValues := []int{0, 0, 0}
	if len(imagesInfo) != 0 {
		var r, g, b uint64
		for i := 0; i < len(imagesInfo); i++ { //über alle Bilder im Pool
			r += uint64(imagesInfo[i].ColorValues[0]) //Farbwerte aufaddieren
			g += uint64(imagesInfo[i].ColorValues[1])
			b += uint64(imagesInfo[i].ColorValues[2])
		}
		averageR := r / uint64(len(imagesInfo)) //Durchschnitt bilden
		averageG := g / uint64(len(imagesInfo))
		averageB := b / uint64(len(imagesInfo))

		colorValues = []int{int(averageR), int(averageG), int(averageB)}
	}

	//Poolinfo auslesen
	pool := groupTy{}
	colPool := dbSession.DB(dbName).C(dbColPools + cookie.Value)
	colPool.Find(bson.M{"name": poolName}).One(&pool)

	//prüfen, ob sich die Werte verändert haben und updaten
	if pool.ColorValues[0] != colorValues[0] ||
		pool.ColorValues[1] != colorValues[1] ||
		pool.ColorValues[2] != colorValues[2] {
		newInfo := pool
		newInfo.ColorValues = colorValues
		colPool.Update(pool, newInfo)
	}
}
