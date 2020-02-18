// imageUtils
/*
 * Diese Klasse stellt Hilfsfunktion für den Umgang mit Bildern zur Verfügung.
 * Darunter fallen Aufgaben wie das Uploaden von Informationen der Bilder, die
 * mittlere Farbe eines Bildes auszurechnen, Bilder skalieren oder Bilder mit
 * gegebener Id aus der Datenbank auszulesen. Hauptsächlich wird die Klasse von
 * imagesAdministration und generators verwendet.
 */
package main

import (
	"image"
	"image/jpeg"
	"net/http"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

/*
 * Erstellt von dem Hochgeladenen Bild Informationen sowie ein Thumbnail und
 * läd diese in die Datenbank hoch. So existiert zu jedem Bild in der DB ein
 * Eintrag in der Info-Collection mit allen notwendigen Informationen.
 */
func uploadImageInformations(img image.Image, imageGridFile *mgo.GridFile,
	imageRGBValues []int, group string, params []string, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)
	//thumbnail erstellen
	dstImage := imaging.Thumbnail(img, 80, 80, imaging.Lanczos)

	//thumbnail in DB speichern
	gridfsThumbs := dbSession.DB(dbName).GridFS(dbGridFsThumbs + cookie.Value)
	gridThumbFile, _ := gridfsThumbs.Create(imageGridFile.Name() + ".Thumb")
	jpeg.Encode(gridThumbFile, dstImage, nil)

	//Infos zum Image erstellen und in DB speichern
	imageInfo := infoTy{
		Id:          imageGridFile.Id().(bson.ObjectId),
		ThumbId:     gridThumbFile.Id().(bson.ObjectId),
		Name:        imageGridFile.Name(),
		Width:       img.Bounds().Max.X,
		Height:      img.Bounds().Max.Y,
		ColorValues: imageRGBValues,
	}

	if group != "all" { //Gruppe wird hinzugefügt
		imageInfo.Groups = []string{group}
	}
	if params != nil { //Parameter für Mosaikerstellung
		imageInfo.Params = params //werden hinzugefügt
	}
	//Infos in Collection hochladen
	colInfo := dbSession.DB(dbName).C(dbColInfo + cookie.Value)
	colInfo.Insert(imageInfo)

	//Close Gridfs files
	gridThumbFile.Close()
}

/*
 * Skalliert die Bilder aus der DB auf die gewünschte Größe und fügt sie
 * anschließend dergewünschten Gruppe hinzu.
 * Ein neues skalliertes Bild wird nur erstellt, wenn es dieses Bild noch nicht
 * in der gewünschten Größe in der Datenbank gibt.
 * Mehrere indentische Bilder in der Datenbank sind somit ausgeschlossen.
 * Die Funktion wird aufgerufen wenn Basismotive skalliert werden oder Bilder
 * zu einem Pool hinzugefügt werden sollen.
 */
func resizeImagesOfDB(imageIds []string, width, height int, group string, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	for i := 0; i < len(imageIds); i++ { //Über alle zu skallierenden Bilder
		objectId := bson.ObjectIdHex(imageIds[i])
		imgExists, imageInfo := checkIfImageExists(objectId, width, height, r)
		//Db collection für die Infos der Bilder
		colInfo := dbSession.DB(dbName).C(dbColInfo + cookie.Value)

		//Neues Bild in gewünschter Größe erstellen
		if !imgExists {
			// zu skallierendes Bild öffnen und skallieren
			img := getImageById(objectId.Hex(), cookie.Value)
			dstImage := scaleImage(img, width, height)

			//skalliertes Bild hochladen
			gridfs := dbSession.DB(dbName).GridFS(dbGridFsBilder + cookie.Value)
			originalFile, _ := gridfs.OpenId(objectId)
			filename := getFileName(originalFile.Name())
			defer originalFile.Close()
			filename = filename + "#" + strconv.Itoa(width) + "X" + strconv.Itoa(height) + ".jpeg"

			//Skalliertes Bild hochladen
			gridNewFile, _ := gridfs.Create(filename)
			jpeg.Encode(gridNewFile, dstImage, nil)
			defer gridNewFile.Close()

			//Informationen zu dem Bild erstellen und hochladen
			newImageInfo := imageInfo
			newImageInfo.Id = gridNewFile.Id().(bson.ObjectId)
			newImageInfo.Name = filename
			newImageInfo.Width = width
			newImageInfo.Height = height
			newImageInfo.Groups = []string{}
			newImageInfo = addImageToGroup(newImageInfo, group)
			colInfo.Insert(newImageInfo)
		} else {
			newImageInfo := addImageToGroup(imageInfo, group)
			colInfo.Update(imageInfo, newImageInfo) //Infos Updaten
		}
	}
}

/*
 * Die Methode prüft, ob eine Version des Bildes in der gefragten Größe bereits
 * vorhanden ist und gibt dessen Informationen zurück.
 * Die Funktion wird beim resizeImagesOfDB zur Prüfung des Datenbestandes genutzt,
 * damit das selbe Bild nicht mehrmal erstellt wird.
 */
func checkIfImageExists(id bson.ObjectId, width, height int,
	r *http.Request) (bool, infoTy) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	//Db collection für die Infos der Bilder
	colInfo := dbSession.DB(dbName).C(dbColInfo + cookie.Value)
	allImageInfos := []infoTy{}
	/*Infos für alle Bilder laden um zu prüfen, ob die geforderte Größe des
	  Bildes schon in der DB existiert.*/
	colInfo.Find(nil).All(&allImageInfos) //Infos für alle Bilder laden

	//Info des Ursprungsbildes Laden
	imageInfo := infoTy{}
	colInfo.Find(bson.M{"_id": bson.M{"$eq": id}}).One(&imageInfo)

	//Prüfen, ob das Bild schon in der richtigen Größe vorhanden ist
	imageExists := false
	for j := 0; j < len(allImageInfos); j++ {
		if allImageInfos[j].ThumbId == imageInfo.ThumbId && allImageInfos[j].Width == width && allImageInfos[j].Height == height {
			imageExists = true
			imageInfo = allImageInfos[j] //Info wird kopiert
			break
		}
	}
	return imageExists, imageInfo
}

/*
 * Die Funktion skalliert ein übergebenes Bild in die gefragte Größe. Dabei
 * werden die Seitenverhältnisse beachtet und falls nötig das größte zentrale
 * Quadrat mit cropCenter ausgeschnitten.
 */
func scaleImage(img image.Image, width, height int) *image.NRGBA {
	/*Bild skallieren - Seitenverhältnisse beachten und zurechtschneiden
	 *mit cropCenter-Methode falls notwendig.*/
	var dstImage *image.NRGBA
	if width == height { //für quadratische Poolkachel
		bounds := img.Bounds()
		if bounds.Max.X == bounds.Max.Y {
			// Bild ist quadratisch - Größe wird nur geändert
			dstImage = imaging.Resize(img, width, height, imaging.Lanczos)
		} else if bounds.Max.X > bounds.Max.Y {
			// Bild ist höher als breit - wird verkleinert und zurechtgeschnitten
			dstImage = imaging.Resize(img, 0, height, imaging.Lanczos)
			dstImage = imaging.CropCenter(dstImage, width, height)
		} else if bounds.Max.X < bounds.Max.Y {
			// Bild ist breiter als hoch - wird verkleinert und zurechtgeschnitten
			dstImage = imaging.Resize(img, width, 0, imaging.Lanczos)
			dstImage = imaging.CropCenter(img, width, height)
		}
	} else {
		//Für Basismotive
		dstImage = imaging.Resize(img, width, height, imaging.Lanczos)
	}
	return dstImage
}

/*
 * Die Funktion berechnet zu dem übergebenen Bild die mittleren r,g,b-
 * Farbwerte.
 * Dazu wird von jedem Pixel des Bildes der r,g,b-Farbanteil ausgelesen und
 * am Ende ein Mittelwert gebildet.
 */
func getImageColorValues(img image.Image) []int {
	//Werte der einzelnen Farbanteile
	var r, g, b, a uint64 //können je nach Bildgröße groß werden
	bounds := img.Bounds()

	//Für jeden Pixel des Bildes werden die Farbanteile ausgelesen.
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			pixelR, pixelG, pixelB, pixelA := img.At(x, y).RGBA()
			r += uint64(pixelR)
			g += uint64(pixelG)
			b += uint64(pixelB)
			a += uint64(pixelA)
		}
	}
	//Formel zur Ausrechnung der durchschnittlichen Farbanteile
	divider := 257 * uint64(bounds.Max.Y) * uint64(bounds.Max.X)
	a = a / divider
	alphaCorrection := 255 / a
	r = r / divider * alphaCorrection
	g = g / divider * alphaCorrection
	b = b / divider * alphaCorrection

	//Slice mit mittleren r,,g,b-Werten wird zurückgegeben
	colorSlice := []int{int(r), int(g), int(b)}
	return colorSlice
}

/*
 * Die Hilfsfunktion gibt zu einer ID in string-Format die Imagedatei zurück.
 */
func getImageById(id string, cookieValue string) image.Image {
	imageID := bson.ObjectIdHex(id)
	gridfsImages := dbSession.DB(dbName).GridFS(dbGridFsBilder + cookieValue)
	gridImageFile, _ := gridfsImages.OpenId(imageID)
	image, _, _ := image.Decode(gridImageFile)
	return image
}

/*
 * Die Hilfsfunktion gibt den Dateinamen zurück, indem die Endung abgeschnitten
 * wird. (bsp .jpeg wird entfernt)
 */
func getFileName(filename string) string {
	splitedExtension := strings.Split(filename, ".")
	splitedSize := strings.Split(splitedExtension[0], "#")
	return splitedSize[0]
}

/*
 * Prüfen, ob die Datei in einem zugelassenen Format formatiert ist, indem die
 * Dateiendung überprüft wird. Die Funktion wird beim Upload benötigt.
 */
func isImage(filename string) bool {
	// Content-Type anhand der Dateiendung bestimmen:
	dateiExt := ""
	if strings.Contains(filename, ".") {
		split2 := strings.Split(filename, ".")
		dateiExt = split2[len(split2)-1]
	} else {
		dateiExt = "unbekannt"
	}
	//Prüfen, ob Endung einem validen Bildformat entspricht
	valid := false
	dateiExt = strings.ToLower(dateiExt)
	if dateiExt == "jpg" || dateiExt == "jpeg" || dateiExt == "png" {
		valid = true
	}
	return valid
}
