// imagesAdministration
/*
 * Diese Klasse verarbeitet alle Nutzeranfragen, die direkt Bilder betreffen.
 * Bilder werden hochgeladen, Informationen ausgelesen, skalliert, ganze Bild-
 * dateien ausgelesen oder komplett gelöscht. Um die Anfragen zu bearbeiten
 * wirden die Hilfsfunktionen aus der Klasse imageUtils verwendet.
 */
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strconv"

	"github.com/globalsign/mgo/bson"
)

//struct um Infos der Bilder an den client zu senden
type filesTy struct {
	Files []infoTy
	Path  string
}

//Struct für Informationen der Bilder in der Datenbank
type infoTy struct {
	Id          bson.ObjectId `bson:"_id"`
	ThumbId     bson.ObjectId `bson:"thumbid"`
	Name        string        `bson:"name"`
	Groups      []string      `bson:"groups"`
	Width       int           `bson:"width"`
	Height      int           `bson:"height"`
	ColorValues []int         `bson:"colorvalues"`
	Params      []string      `bson:"prams"` //Parameter für generierte Mosaiken
} //(nCount, ReuseImages, Poolname)

//Struct um eine Slice an Images aus einer Request zu lesen und auf die
//geforderte Größe zu skallieren.
type resizeRequest struct {
	ImageIDs []string
	Width    string
	Height   string
	Group    string
}

/*
 * Die ausgewählten Bilder werden ausgelesen. Das Format der Bilder wird anhand
 * ihrer Dateiendungen überprüft. Jedes Bild wird nacheinander in die DB(GridFS)
 * geladen. Wenn ein Bild direkt in eine Gruppe geladen wird, wird es dieser
 * zugewiesen. Wenn es sich dabei um einen Pool handelt, wird ein weiteres in
 * der korrekten Größe skalliertes Bild erstellt und dem Pool hinzugefügt.
 *
 */
func uploadImagesHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	// GridFs-collection "bilder" dieser DB:
	gridfsImages := dbSession.DB(dbName).GridFS(dbGridFsBilder + cookie.Value)

	err = r.ParseMultipartForm(2000000) // bytes
	filesUploaded := 0                  //counter für hochgeladene Bilder
	filesFailed := 0                    //counter für Bilder im falschen Format

	if err == nil { // => alles ok
		formdataPointer := r.MultipartForm

		if formdataPointer != nil { // beim ersten request ist die Form leer!
			files := formdataPointer.File["inputFiles"]

			for i, _ := range files {
				// Datei hochladen, wenn es ein Bild ist
				if isImage(files[i].Filename) {
					// upload-files öffnen:
					uplFile, err := files[i].Open()
					check_ResponseToHTTP(err, w)
					defer uplFile.Close()

					img, _, err := image.Decode(uplFile) //file in Image konvertieren
					check_ResponseToHTTP(err, w)

					/* Prüfen, ob Bilder direkt in einen Pool geladen werden.
					 * Falls ja, wird die Größe auf die Poolgröße geändert und
					 * das Bild dem Pool hinzugefügt.*/
					groupName := r.FormValue("groupName") //Gruppenname wird ausgelesen
					groupTyp := r.FormValue("groupTyp")
					if groupTyp == "Pools" {

						pool := groupTy{} //Poolgröße auslesen
						dbSession.DB(dbName).C(dbColPools + cookie.Value).Find(
							bson.M{"name": bson.M{"$eq": groupName}}).One(&pool)

						// grid-file mit diesem Namen erzeugen:
						filename := getFileName(files[i].Filename)
						filename = filename + "#" + strconv.Itoa(pool.Size) +
							"X" + strconv.Itoa(pool.Size)
						gridImageFile, _ := gridfsImages.Create(filename)

						//Bild scalieren und hochladen
						dstImage := scaleImage(img, pool.Size, pool.Size)
						jpeg.Encode(gridImageFile, dstImage, nil)
						gridImageFile.Close()

						imgFile, _ := gridfsImages.Open(filename)
						img, _ = jpeg.Decode(imgFile)
						defer imgFile.Close()

						colorValues := getImageColorValues(img)
						uploadImageInformations(img, imgFile, colorValues, groupName, nil, r)
					} else {
						//Gridfile mit namen erzeugen
						gridImageFile, err := gridfsImages.Create(files[i].Filename)
						defer gridImageFile.Close()

						//zum Anfang des Files springen
						_, err = uplFile.Seek(0, 0)
						check_ResponseToHTTP(err, w)

						// ImageFile in GridFSkopieren:
						jpeg.Encode(gridImageFile, img, nil)
						check_ResponseToHTTP(err, w)

						//notwendige Informationen erzeugen und diese hochladen
						imageRGBValues := getImageColorValues(img)
						uploadImageInformations(img, gridImageFile, imageRGBValues,
							"all", nil, r)
					}
					filesUploaded++
				} else {
					filesFailed++
				}
			}
		}
	}

	//Rückmeldung über Erfolg an den Client senden
	msg := ""
	if filesFailed == 0 && filesUploaded != 0 {
		msg = fmt.Sprintf("%d Bilder wurden erfolgreich hochgeladen!", filesUploaded)
	} else if filesFailed != 0 {
		msg = fmt.Sprintf("Es wurden %d Bilder hochgeladen. %d Dateien im falschen Format.",
			filesUploaded, filesFailed)
	} else {
		msg = "Keine Datei ausgewählt!"
	}
	sendMessageToClient(w, msg)
}

/*
 * Handler zum Auslesen der Thumbnails zu der ausgewählten Gruppe.
 * Gibt eine Json Datei mit den IDs der Thumbnails der erfragten Bilder und
 * den Speicherort zurück. Diese werden beim Client im Preview den Thumbnails
 * zugewiesen.
 */
func getImagesHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	r_selectedCollection := r.FormValue("selected") //ausgewählte Collection wird ausgelesen

	gridfsImages := dbSession.DB(dbName).C(dbColInfo + cookie.Value)

	allFiles := []infoTy{}
	gridfsImages.Find(nil).All(&allFiles) //alle Bilder werden ausgelesen

	/*Es wird überprüft, ob die Bilder der Gruppe angehören und die ID des
	  dazugehörigen Thumbnails ausgelesen und dem Clienten übergeben*/
	selectedFiles := []infoTy{}
	if r_selectedCollection == "all" {
		selectedFiles = allFiles
	} else {
		for i := 0; i < len(allFiles); i++ { //für jedes Bild

			for j := 0; j < len(allFiles[i].Groups); j++ { //werden alle Gruppennamen durchgegangen
				if allFiles[i].Groups[j] == r_selectedCollection {
					matchedFile := allFiles[i]
					selectedFiles = append(selectedFiles, matchedFile)
				}
			}
		}
	}

	//Bilderinfos werden dem Sendestruct übergeben
	sendFiles := filesTy{
		Files: selectedFiles,
		Path:  dbGridFsThumbs,
	}
	//format to json and send it
	jData, _ := json.Marshal(sendFiles)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

/*
 * Der Handler sendet die einzelnen Bilder an den Client.
 * In einer URL werden ID und GridFS-Name übergeben, mit denen auf das passende
 * Bild in der DB zugegriffen wird
 */
func getImageHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	// request lesen:
	r_gridfsName := r.URL.Query().Get("gridfsName")
	r_fileId := r.URL.Query().Get("fileId")
	objId := bson.ObjectIdHex(r_fileId) //Id aus übergebenen String erstellen

	// angefordertes Bild der geforderten Collection öffnen
	gridfs := dbSession.DB(dbName).GridFS(r_gridfsName + cookie.Value)
	gridFile, err := gridfs.OpenId(objId)

	// image senden:
	w.Header().Add("Content-Type", "image/jpeg")
	_, err = io.Copy(w, gridFile)
	check_ResponseToHTTP(err, w)

	err = gridFile.Close()
	check_ResponseToHTTP(err, w)
}

/*
 * Der Handler liest die zur Skallierung eines Bildes notwendigen Daten aus
 * und ruft die resizeImages-Funktion auf.
 */
func resizeImagesHandler(w http.ResponseWriter, r *http.Request) {
	//Daten aus JSON Request decodieren
	data := resizeRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&data)

	//Werte in Int konvertieren
	heightVal, _ := strconv.Atoi(data.Height)
	widthVal, _ := strconv.Atoi(data.Width)

	resizeImagesOfDB(data.ImageIDs, widthVal, heightVal, data.Group, r)

	//Rückmeldung über Erfolg an den Client senden
	sendMessageToClient(w, "Bild erfolgreich skalliert!")
}

/*
 * Der Handler sucht die Informationen zu einem angefragten Bild aus der
 * DB und sendet sie an den Client.
 */
func getImageInfoHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	imageId := bson.ObjectIdHex(r.FormValue("id")) //id auslesen

	//Infomationen aus DB-Collection lesen
	info := infoTy{}
	colInfo := dbSession.DB(dbName).C(dbColInfo + cookie.Value)
	colInfo.Find(bson.M{"_id": bson.M{"$eq": imageId}}).One(&info)

	//format to json and send it
	jData, _ := json.Marshal(info)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

/*
 * Dieser Handler löscht die ausgewählten Bilder und deren Infos aus der DB.
 * Es wird überprüft, ob die Thumbnails der Bilder noch benötigt werden. Falls
 * nicht, wird auch das Thumbnail zu dem Bild gelöscht.
 */
func removeImagesHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	//Daten aus JSON Request decodieren
	data := groupRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&data)

	//Informationen zu allen Bildern auslesen
	colInfo := dbSession.DB(dbName).C(dbColInfo + cookie.Value)
	allImageInfos := []infoTy{}
	colInfo.Find(nil).All(&allImageInfos)

	infoRemovedImage := infoTy{}
	//Für alle zu löschenden Bilder
	for i := 0; i < len(data.ImageIDs); i++ {
		//thumbnailId des zu löschenden Bildes auslesen
		objectId := bson.ObjectIdHex(data.ImageIDs[i])
		colInfo.Find(bson.M{"_id": bson.M{"$eq": objectId}}).One(&infoRemovedImage)

		//Prüfen, ob ein weiteres Bild ebenfalls einen Verweis auf das Thumbnail hat
		counter := 0
		thumbNeeded := false
		for j := 0; j < len(allImageInfos); j++ {
			if infoRemovedImage.ThumbId == allImageInfos[j].ThumbId {
				counter++
				if counter == 2 { //2 da der eigene Verweis auch dabei ist
					break
				}
			}
		}
		if counter == 2 {
			thumbNeeded = true //Thumbnail wird noch von einem anderen
		} //Bild verwendet

		//Bild und Bildinfo komplett löschen
		if data.Group == "all" {
			colInfo.RemoveId(objectId)
			colInfo.Find(nil).All(&allImageInfos)
			gridfsImage := dbSession.DB(dbName).GridFS(dbGridFsBilder + cookie.Value)
			gridfsImage.RemoveId(objectId)
			//Thumbnail wird entfernt, wenn kein weiteres Bild es verwendet
			if !thumbNeeded {
				gridfsThumb := dbSession.DB(dbName).GridFS(dbGridFsThumbs + cookie.Value)
				gridfsThumb.RemoveId(infoRemovedImage.ThumbId)
			}
		} else { //Bild wird aus der gewählten Gruppe entfernt
			newImageInfo := infoRemovedImage
			for j := 0; j < len(newImageInfo.Groups); j++ {
				if newImageInfo.Groups[j] == data.Group {
					newImageInfo.Groups = append(newImageInfo.Groups[:j], newImageInfo.Groups[j+1:]...)
					colInfo.Update(infoRemovedImage, newImageInfo) //Gruppen updaten
				}
			}
		}
	}
	//Rückmeldung über Erfolg an den Client senden
	sendMessageToClient(w, "Löschen erfolgreich!")
}
