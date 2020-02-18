// generators
/*
 * Diese Klasse stellt die beiden Generatoren der Seite zur Verfügung. Es können
 * zum einen Bilder generiert werden und zum anderen das Mosaikfoto generiert
 * werden. Das generierte Mosaikbild kann zusätzlich heruntergeladen werden. Für
 * die Bearbeitung der verwendeten Bilder werden Funktionen der Klasse imageUtils
 * verwendet.
 */
package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/globalsign/mgo/bson"
)

/*Struct um die Werte des Farbabgleichs zwischen einem Pixel und einer Kachel
 *abzuspeichern. Nach dem Difference-Wert werden die Kacheln sortiert um später
 *mit der Id auf das Bild zugreifen zu können. Die Placevariable ermöglicht den
 *Zugriff auf das Ausgangsobjekt, um dies evtl für weitere Suchen bei
 *nicht erlaubter Mehrfachverwendung zu sperren.
 */
type colorDifferenceTy struct {
	Id         bson.ObjectId
	Difference float64
	Place      int
}

/*Struct für Farbabgleich der Kacheln in dem festgelegt werden kann,
  ob die Kachel schon genutz wurde. Wird benötigt, falls Mehrfachnutzung
  der Kacheln nicht erlaubt ist.*/
type genInfoTy struct {
	Id          bson.ObjectId `bson:"_id"`
	Size        int           `bson:"width"`
	ColorValues []int         `bson:"colorvalues"`
	Used        bool          `bson:"used"`
}

/*
 * Der Handler erstellt einen Pool mit generierten Bildern. Die Übergabeparameter
 * (Größe, Anzahl, mit Rechteck, Rechteckgröße) legen fest, wie die Bilder
 * generiert werden sollen. Es werden Bilder in zufälliger Farbe mit einem
 * weißen Punkt in der Mitte erstellt. Zusätzlich ist es möglich Rechtecke in
 * einer weiteren zufälligen Farbe und gewünschter Größe zu erstellen.
 */
func generateImagesHandler(w http.ResponseWriter, r *http.Request) {
	//Parameter der Erstellung auslesen
	poolName := r.FormValue("genPoolName")
	tilesSize := r.FormValue("tileSize")
	tilesCount, _ := strconv.Atoi(r.FormValue("tileCount"))
	setRectStr := r.FormValue("rect")
	rectSizeStr := r.FormValue("rectSize")

	//Neuen Pool erstellen und Parameter der Erstellung abspeichern
	params := []string{setRectStr, rectSizeStr}
	success := createNewGroup(w, "pool", poolName, tilesSize, r, params)

	if success {
		//Über die Anzahl der geforderten Bilder iterieren
		for i := 0; i < tilesCount; i++ {
			//zufällige Farben erstellen
			var colorValues1 [3]uint8
			var colorValues2 [3]uint8
			for j := 0; j < len(colorValues1); j++ {
				colorValues1[j] = uint8(rand.Intn(255))
				colorValues2[j] = uint8(rand.Intn(255))
			}
			randColor := color.RGBA{colorValues1[0], colorValues1[1], colorValues1[2], 255}
			randColor2 := color.RGBA{colorValues2[0], colorValues2[1], colorValues2[2], 255}
			colorA := color.RGBA{255, 255, 255, 255}

			//Bild mit zufälliger Farbe erstellen
			size, _ := strconv.Atoi(tilesSize)
			rect := image.Rect(0, 0, size, size)
			img := image.NewRGBA(rect)

			//Prüfen ob und wie die Rechtecke gesetzt werden sollen
			setRect := false
			if setRectStr == "on" {
				setRect = true
			}
			rectSize, _ := strconv.Atoi(rectSizeStr)
			//Einzelne Punkte setzen
			for x := 0; x < size; x++ {
				for y := 0; y < size; y++ {
					//setze ein Rechteck in der gefragten Größe
					if setRect && x < rectSize && y < rectSize {
						img.Set(x, y, randColor)
						//setze einen Weißen Punkt in die Mitte des Bildes
					} else if x > size/2-1 && x < size/2+1 && y > size/2-1 && y < size/2+1 {
						img.Set(x, y, colorA)
						//setze den Rest des Bildes
					} else {
						img.Set(x, y, randColor2)
					}
				}
			}
			//Namen festlegen und Bildinformationen erstellen sowie hochladen
			imgName := poolName + "_" + tilesSize + "_" + strconv.Itoa(i) + ".jpg"
			uploadGeneratedImages(img, imgName, poolName, nil, r)
		}
	}
}

/*
 * Generierte Bilder werden formatiert und in die DB geladen. Die Funktion wird
 * von generateImagesHandler und von generateMosaicHandler aufgerufen.
 */
func uploadGeneratedImages(img *image.RGBA, filename, group string, params []string, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	buf := new(bytes.Buffer)
	jpeg.Encode(buf, img, nil) //Bild in Bytes kodieren

	//Bild in die DB laden
	gridfs := dbSession.DB(dbName).GridFS(dbGridFsBilder + cookie.Value)
	gridfsFile, _ := gridfs.Create(filename)
	defer gridfsFile.Close()
	jpeg.Encode(gridfsFile, img, nil)

	//Notwendige Informationen erstellen und in die DB laden
	imageRGBValues := getImageColorValues(img)
	uploadImageInformations(img, gridfsFile, imageRGBValues, group, params, r)
}

/*
 * Der Handler erstellt von einem ausgewählten Basismotiv und einem ausgewählten
 * Pool mit den eingestellten Optionen ein Mosaikbild.
 * Dazu werden alle Bilder des Pools ausgelesen, über diese iteriert und von
 * jedem Pixel des Basismotives der Farbabstand zu jeder Kachel des Pool
 * berechnet. Je nach parametisierter Einstellung wird dann die passende Kachel
 * anstelle des Pixels gesetzt.
 */
func generateMosaicHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	//Formdaten auslesen
	bestFitStr := r.FormValue("bestFit")
	countNStr := r.FormValue("countBestFit")
	motiveIdStr := r.FormValue("motiveId")
	poolName := r.FormValue("poolName")
	reuseImagesStr := r.FormValue("reuseImages")
	//Bool konvertieren
	reuseImages := false
	bestFit := false
	if reuseImagesStr == "on" {
		reuseImages = true
	}
	if bestFitStr == "on" {
		bestFit = true
	}

	//Basismotiv auslesen in Image konvertieren
	motiveInfo := infoTy{}
	motiveImage := getImageById(motiveIdStr, cookie.Value)
	colInfo := dbSession.DB(dbName).C(dbColInfo + cookie.Value)
	colInfo.Find(bson.M{"_id": bson.M{"$eq": bson.ObjectIdHex(motiveIdStr)}}).One(&motiveInfo)

	//Informationen der Kacheln des Pools auslesen
	tilesInfo := []genInfoTy{}
	colInfo.Find(bson.M{"groups": poolName}).All(&tilesInfo)
	tileSize := tilesInfo[0].Size

	//Mosaik Image erstellen
	motiveBounds := motiveImage.Bounds()
	mosaicImage := image.NewRGBA(image.Rect(0, 0,
		motiveBounds.Max.X*tileSize, motiveBounds.Max.Y*tileSize))
	imagePoint := image.Point{0, 0}

	//Prüfen, ob genug Kacheln im Pool sind (bei Nicht-Mehrfachverwendung)
	countPixel := motiveBounds.Max.X * motiveBounds.Max.Y
	if reuseImages || (!reuseImages && countPixel < len(tilesInfo)) {
		//jeden Pixel des Motivs durchlaufen und mit Kacheln des Pools abgleichen
		for x := 0; x < motiveBounds.Max.X; x++ {
			for y := 0; y < motiveBounds.Max.Y; y++ {
				//Farbwerte eines Pixels des Basismotivs auslesen
				pixelR, pixelG, pixelB, _ := motiveImage.At(x, y).RGBA()
				pixelR /= 257
				pixelG /= 257
				pixelB /= 257

				//Farbliche Differenz zu jeder Kachel berechnen und in Slice speichern
				colDifSlice := []colorDifferenceTy{}
				for i := 0; i < len(tilesInfo); i++ {
					if reuseImages || (!reuseImages && !tilesInfo[i].Used) {
						difR := uint32(tilesInfo[i].ColorValues[0]) - pixelR
						difG := uint32(tilesInfo[i].ColorValues[1]) - pixelG
						difB := uint32(tilesInfo[i].ColorValues[2]) - pixelB
						colorDif := math.Sqrt(float64(difR*difR + difG*difG + difB*difB))
						//speichere den Wert der Abweichung mit der Id in der Slice
						tileInfo := colorDifferenceTy{
							Id:         tilesInfo[i].Id,
							Difference: colorDif,
							Place:      i,
						}
						colDifSlice = append(colDifSlice, tileInfo)
					}
				}
				//sortiere die Slice
				sort.Slice(colDifSlice, func(i, j int) bool {
					return colDifSlice[i].Difference < colDifSlice[j].Difference
				})
				//passende Kachel auswählen nach übergebenen Parametern
				var choice int
				if bestFit {
					choice = 0
				} else {
					n, _ := strconv.Atoi(countNStr)
					choice = rand.Intn(n)
				}
				//Id auslesen und Kachel als verwendet markieren
				tileId := colDifSlice[choice].Id.Hex()
				tilesInfo[colDifSlice[choice].Place].Used = true

				//Image der passenden Kachel auslesen und im Mosaik setzen
				tileImage := getImageById(tileId, cookie.Value)
				tileRect := image.Rect(x*tileSize, y*tileSize, x*tileSize+tileSize, y*tileSize+tileSize)
				draw.Draw(mosaicImage, tileRect, tileImage, imagePoint, draw.Src)
			}
		}
		//Parameter für die Generierung werden zum speichern übergeben
		params := []string{countNStr, reuseImagesStr, poolName}
		mosaicName := getFileName(motiveInfo.Name)
		mosaicName = mosaicName + "_" + poolName + "_" + params[0] + params[1] + ".jpeg"
		//Mosaik wird konvertiert und hochgeladen
		uploadGeneratedImages(mosaicImage, mosaicName, "mosaics", params, r)
		sendMessageToClient(w, "Mosaik erfolgreich erstellt!")
	} else {
		sendMessageToClient(w, "Zu wenig Kacheln im Pool!")
	}
}

/*
 * Der Handler downloaded das gefragte Mosaikbild in den downloads-Ordner.
 */
func downloadMosaicHandler(w http.ResponseWriter, r *http.Request) {
	//cookie auslesen
	cookie, _ := r.Cookie(cookieName)

	id := bson.ObjectIdHex(r.FormValue("id")) //id aus Form auslesen

	//Bild aus DB öffnen
	imagesGridfs := dbSession.DB(dbName).GridFS(dbGridFsBilder + cookie.Value)
	imageFile, _ := imagesGridfs.OpenId(id)
	defer imageFile.Close()

	//Neue Datei im Downloadordner erstellen
	downloadFile, _ := os.Create("downloads/" + imageFile.Name())
	defer downloadFile.Close()

	//Bilddatei in Downloadfile kopieren
	io.Copy(downloadFile, imageFile)
}
