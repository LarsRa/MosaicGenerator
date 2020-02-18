    /*
     * Die Funktion erstellt einen neuen Pool mit generierten Bildern. 
     * Generierungsparameter entscheiden, ob die Bilder mit Rechtecken erstellt werden
     * und wie groß diese sein sollen. 
     * Die Request wird an den generateImages Handler gesendet.
     */
    function generateImages() {
        var formData = new FormData(document.getElementById("generatePool"));
        var tileSize = formData.get("tileSize");
        var tilesCount = formData.get("tileCount");
        var poolName = formData.get("genPoolName");
        var rect = formData.get("rect");
        var rectSize = formData.get("rectSize");
  
        //Eingaben werden validiert - kein Feld darf leer sein
        if (tileSize != "" && tilesCount != ""
          && poolName != "" && poolName != "all" && poolName != "mosaics"
          && (!rect || rect && rectSize != "")) {
  
          //prüfen, ob die eingegebenen Größen passen
          if (rectSize <= tileSize / 2) {
            //Eingaben leeren
            document.getElementById("generatePool").reset();
            document.getElementById("rectSize").disabled =true;
            //Request erstellen und generateImages-Handler aufrufen
            var xhr = new XMLHttpRequest();
            xhr.open("POST", "http://localhost:4242/generateImages");
            xhr.onload = function () {
              var response = JSON.parse(xhr.responseText);
              getAllGroups();
              getImages("mosaics");
              showMessage(response.Text);
            }
            xhr.send(formData);
          } else {
            showMessage("Das Rechteck darf maximal halb so groß wie die Kachel sein.");
          }
  
        } else {
          showMessage("Eingabe nicht gültig. Es müssen alle Felder gesetzt werden.");
        }
      }
  
      /*
       * Die Methode überprüft, ob die Eingaben zur Generierung eines Mosaiks valide sind.
       * Wenn die Eingaben valide sind, werden die notwendigen Daten per JSON request an 
       * den Server-Handler /generateMosaic gesendet. Falls die Daten nicht valide sind,
       * wird eine Fehlernachricht im Header angezeigt.
       */
      function generateMosaic() {
        //Prüfen, ob nur ein Bild ausgewählt ist
        var selectedImages = getSelectedIDs();
        if (selectedImages.length == 1 && selectedImages[0] != 0) {
          //Input-Werte auslesen
          var strArray = document.getElementById("genMosaicPoolName").value.split(" ");
          var poolName = strArray[1];
          var bestFit = document.getElementById("bestFitCheckbox").checked;
          var countBestFit = document.getElementById("countBestFit").value;
  
          //Prüfen, ob input für Bestfit und N valide ist
          if (bestFit || (!bestFit && countBestFit > 1 && countBestFit < 100)) {
            var data = new FormData(document.getElementById("mosiacGenerator"));
            data.set("motiveId", selectedImages[0]);
            data.set("poolName", poolName);
  
            //Request erstellen und an den generateMosaic-Handler senden
            var xhr = new XMLHttpRequest();
            xhr.open("POST", "http://localhost:4242/generateMosaic");
            xhr.onload = function () {
              var response = JSON.parse(xhr.responseText);
              showMessage(response.Text);
              getAllGroups();
              getImages("mosaics");
              //Eingaben leeren
              document.getElementById("mosiacGenerator").reset();
              document.getElementById("countBestFit").disabled = false;
            }
            xhr.send(data);
          } else {
            showMessage("Methode Bestfit oder N muss zwischen 2 und 100 gesetzt werden");
          }
        } else {
          showMessage("Es muss genau ein Bild für die Generierung ausgewählt sein.");
        }
      }
      
    /*
     * Die Funktion liest aus der Source des dargestellten Bildes die ID aus und
     * sendet diese an den downloadMosaic-Handler.
     */
    function downloadMosaic() {
        //ID wird aus Bildsource ausgelesen
        var imageSrc = document.getElementById("showImage").getAttribute("src");
        var strArray = imageSrc.split("=");
        var id = strArray[strArray.length - 1];
  
        var formdata = new FormData();
        formdata.set("id", id);                 //Id wird in Formdata gesetzt
  
        //Request an downloadMosaic-Handler senden
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "http://localhost:4242/downloadMosaic");
        xhr.onload = function () {
          showMessage("Download erfolgreich. Datei in /downloads/");
        }
        xhr.send(formdata);
      }