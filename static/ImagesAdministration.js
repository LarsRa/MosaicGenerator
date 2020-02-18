    /*
     * Ausgewählte Dateien werden ausgelesen und an den uploadImages-Handler gesendet.
     * Die Antwort wird im Header dargestellt und der Input geleert.
     */
    function upLoadImages() {
        var formData = new FormData(document.getElementById("uploadForm"));
  
        //Gruppennamen auslesen, in die die Bider geladen werden sollen
        var selectorPreview = document.getElementById("selectShown");
        var strArray = selectorPreview.value.split(" ");
        var groupTyp = strArray[0];
        var groupName = strArray[1];
        if(groupName=="mosaics"){
          groupName = "all";
        }
        formData.set("groupTyp",groupTyp);
        formData.set("groupName",groupName);
  
        //Request erstellen uploadImages-Handler aufrufen
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "http://localhost:4242/uploadImages");
        xhr.onload = function () {
          //JSON response parsen
          var response = JSON.parse(xhr.responseText);
  
          showMessage(response.Text);                             //Nachricht über Erfolg/Misserfolg anzeigen       
          getImages();                                            //Bildervorschau neu laden
  
          document.getElementById("inputFiles").value = null;     //input leeren
        }
        xhr.send(formData);
      }
  
      /*Ausgewählte Bilder werden aus der ausgewählten Gruppe gelöscht*/
      function removeImages() {
        var imageIDs = getSelectedIDs();
        var strArray = document.getElementById("selectShown").value.split(" ");
        var data = JSON.stringify({ "ImageIDs": imageIDs, "Group": strArray[1] });
  
        //Request erstellen removeImages-Handler aufrufen
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "http://localhost:4242/removeImages");
        xhr.onload = function () {
          //JSON response parsen
          var response = JSON.parse(xhr.responseText);
          showMessage(response.Text);
          getImages();
        }
        xhr.send(data);
      }
  
      /*Bilder werden aus der Datenbank geladen und in der Vorschaun angezeigt*/
      function getImages(name) {
        var selector = document.getElementById("selectShown");
        var formdata = new FormData();
        //gewünschte Gruppe auslesen (übergeben oder aus selector)
        if (typeof (name) == "string") {
          formdata.set("selected", name);
        } else {
          var strArray = selector.value.split(" ");
          var selected = strArray[1];
          formdata.set("selected", selected);
        }
        //Request erstellen getImages-Handler aufrufen
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "http://localhost:4242/getImages");
        xhr.onload = function () {
          //JSON response parsen
          var response = JSON.parse(xhr.responseText);
          var container = document.getElementById("thumbnailContainer");
  
          //Container mit Thumbails leeren
          while (container.firstChild) {
            container.removeChild(container.firstChild);
          }
  
          //Container mit neuen Thumbnails füllen
          if (response.Files.length == 0) {
            var text = document.createElement("p");
            text.innerText = "Es sind keine Bilder vorhanden!";
            container.appendChild(text);
          } else {
            //einzelne Thumbnails erstellen
            for (var i = 0; i < response.Files.length; i++) {
              var thumbnail = document.createElement("img");
              thumbnail.addEventListener("click", thumbClicked);
              thumbnail.setAttribute("id", response.Files[i].Id);
              thumbnail.addEventListener("mouseenter", showSize);         //Größe wird angezeigt beim Hovern
              thumbnail.addEventListener("mouseleave", showSize);
              thumbnail.style.position = "relative";
              thumbnail.style.textAlign = "center";
              thumbnail.style.color = "red";
              thumbnail.setAttribute("src", "/getImage?fileId=" + response.Files[i].ThumbId + "&gridfsName=" + response.Path);
              thumbnail.className = "thumbnail";
              container.appendChild(thumbnail);
            }
          }
        }
        xhr.send(formdata);
      }
  
      /*Zeigt die Größe des Bildes an, wenn über das Thumbnail gehovert wird*/
      function showSize() {
        //Sizetext befindet sich über dem Thumbnailviewer
        var sizeText = document.getElementById("sizeText");
        if (sizeText.hidden) {
          var id = this.getAttribute("id");       //id des Thumbnails wird ausgelesen
          var data = new FormData();
          data.set("id", id);
  
          //Request erstellen und de getImageInfo-Handler aufrufen
          var xhr = new XMLHttpRequest();
          xhr.open("POST", "http://localhost:4242/getImageInfo");
          xhr.onload = function () {
            //JSON response parsen
            var response = JSON.parse(xhr.responseText);
            //Größe des Bildes setzen 
            sizeText.innerText = "Größe: " + response.Width + " X " + response.Height;
          }
          xhr.send(data);
        }
        sizeText.hidden = !sizeText.hidden;
      }
  
      /*
       * Nachdem ein thumbnail ausgewählt wurde, wird die Klasse des Elements geändert
       * und das Bild in Originalgröße angezeigt.
       */
      function thumbClicked() {
        if (this.className == "thumbnail") {
          this.className = "thumbnail selectedThumb";             //Klasse mit Umrandung setzen
          var image = document.getElementById("showImage");
          //Bild in Originalgröße anzeigen
          image.setAttribute("src", "/getImage?" + "gridfsName=IMAGES&fileId=" + this.getAttribute("id"));
          getImageInformation(this.getAttribute("id"));
        } else {
          this.className = "thumbnail";                           //Klasse mit Umrandung aufheben
        }
      }
          /*
     * Die Funktion skalliert das ausgewählte Bild auf die gewünschte Größe. Das skallierte Bild
     * darf maximal 1600 px groß sein. Die Anfrage wird an den resizeImages handler gesendet
     *
     */
    function resizeImage() {
        //werte auslesen
        var width = document.getElementById("width");
        var height = document.getElementById("height");
        var imageIDs = getSelectedIDs();
  
        //Prüfen ob eingegebene Werte valide sind
        if (imageIDs.length == 1) {
          if (width.value > 0 && width.value <= 1600 && height.value > 0 && height.value <= 1600) {
            var strArray = document.getElementById("selectShown").value.split(" ");
            var group;
            if (strArray[0]!="Pools"){
              group = strArray[1];
            }else{
              group = "all";
            }
            var data = JSON.stringify({ 
              "ImageIDs": imageIDs, 
              "Width": width.value, 
              "Height": height.value, 
              "Group":group
            });
            //Request erstellen und resizeImages-Handler aufrufen
            var xhr = new XMLHttpRequest();
            xhr.open("POST", "http://localhost:4242/resizeImages");
            xhr.onload = function () {
              //JSON response parsen
              var response = JSON.parse(xhr.responseText);
              showMessage(response.Text);
              getImages();
            }
            xhr.send(data);
  
            //Felder leeren
            width.value = "";
            height.value = "";
  
          } else {
            showMessage("Die Größe muss 1px - 1600px betragen!");
          }
        } else {
          showMessage("Es darf nur ein Bild ausgewählt sein!");
        }
      }

          /*
     * Es werden die Informationen zu einem ausgewählten Bild abgerufen und
     * die passenden Elemente in der Infobox mit Dom-Scripting erzeugt.
     */
    function getImageInformation(id) {
        var data = new FormData();
        data.set("id", id);
        //Request erstellen und getInformation-Handler aufrufen
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "http://localhost:4242/getImageInfo");
        xhr.onload = function () {
          //JSON response parsen
          var response = JSON.parse(xhr.responseText);
  
          //Container leeren
          var infoContainer = document.getElementById("imageInformation");
          removeAllChilds(infoContainer);
          //Überschrift hinzufügen
          var heading = document.createElement("h3");
          heading.innerText = "Bildinformationen:";
          infoContainer.appendChild(heading);
          //name hinzufügen
          var name = document.createElement("p");
          name.innerText = "Name: " + response.Name;
          infoContainer.appendChild(name);
          //Größe hinzufügen
          var size = document.createElement("p");
          size.setAttribute("id", "imageSize");
          size.innerText = "Größe: " + response.Width + " x " + response.Height;
          infoContainer.appendChild(size);
          //Farbwerte hinzufügen
          var rgb = document.createElement("p");
          rgb.innerText = "Farbwerte: " + response.ColorValues[0] + " "
            + response.ColorValues[1] + " " + response.ColorValues[2];
          infoContainer.appendChild(rgb);
  
          //für Mosaiken Parameter der Erstellung hinzufügen
          if (response.Params.length != 0) {
            var paramsHeading = document.createElement("h4");       //Überschrift
            paramsHeading.innerText = "Generierungsparameter:";
            infoContainer.appendChild(paramsHeading);
  
            var poolName = document.createElement("p");             //Poolname
            poolName.innerText = "Poolname: " + response.Params[2];
            infoContainer.appendChild(poolName);
  
            var reuseTiles = document.createElement("p");           //Mehrfachverwendung der Kacheln
            reuseTiles.innerText = "Kacheln mehrfach verwendet: " + response.Params[1];
            infoContainer.appendChild(reuseTiles);
  
            var bestFit = document.createElement("p");              //Bestfit
            if (response.Params[0] == "") {
              bestFit.innerText = "Bestfit: true";
              infoContainer.appendChild(bestFit);
            } else {
              bestFit.innerText = "Bestfit: false";
              infoContainer.appendChild(bestFit);
              var nCount = document.createElement("p");             //Anzahl der Auswahl der besten Kachel
              nCount.innerText = "Kachel aus den besten " + response.Params[0] + " Kacheln ausgewählt.";
              infoContainer.appendChild(nCount);
            }
            //Downloadbutton hinzufügen
            var downloadBtn = document.createElement("button");
            downloadBtn.innerText = "Download";
            downloadBtn.className = "btn formButton";
            downloadBtn.addEventListener("click", downloadMosaic);
            infoContainer.appendChild(downloadBtn);
  
            //dazugehörige Poolinformationen laden
            getPoolInformation(response.Params[2]);
          }
        }
        xhr.send(data);
      }