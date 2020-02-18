    /*
     * Die Funktion erstellt eine Gruppe (Pool oder Sammlung), indem sie die Kategorie sowie
     * den Namen der Gruppe ausliest und an den createGroup-Handler sendet. 
     * Wenn die Gruppe erfolgreich erstellt wurde, werden die ausgewählten Bilder der Gruppe 
     * mit der Methode addToGroup() hinzugefügt.
     */
    function createNewGroup(category) {
        //Passende Formdaten für Pool oder Sammlung auslesen
        if (category == "pool") {
          // Formdaten auslesen und Kategorie hinzufügen
          var formData = new FormData(document.getElementById("createPoolForm"));
          formData.set("category", "pool");
  
          //Eingabe in die Inputfelder löschen
          document.getElementById("poolName").value = "";
          document.getElementById("size").value = "";
        } else {
          // Formdaten auslesen und Kategorie hinzufügen
          var formData = new FormData(document.getElementById("createCollectionForm"));
          formData.set("category", "collection");
  
          //Eingabe in die Inputfelder löschen
          document.getElementById("collectionName").value = "";
        }
  
        var groupName = formData.get("name");   //wird benötigt, um die ausgewählten Bilder in onload hinzuzufügen
  
        //Request erstellen und createGroup-Handler aufrufen
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "http://localhost:4242/createGroup");
        xhr.onload = function () {
          //JSON response parsen
          var response = JSON.parse(xhr.responseText);
          showMessage(response.Text);
          getAllGroups();                               //Gruppen neu laden
  
          //Ausgewählte Bilder auslesen und der erstellten Gruppe hinzufügen
          var imageIDs = getSelectedIDs();
          if (imageIDs.length != 0) {
            var selector;
            if (category == "pool") {
              selector = document.getElementById("selectPool");
            } else {
              selector = document.getElementById("selectCollection");
            }
            //Bilder Gruppe hinzufügen
            addToGroup(selector.getAttribute("id"), groupName);
          }
        }
        xhr.send(formData);
      }
  
      /*
       * Die Methode sendet eine Request an den getAllGroups Handler, um alle für den Nutzer 
       * verfügbaren Gruppen zu erhalten. Folgend werden alle vorhandenen Selectoren geleert
       * und die empfangenden Gruppen den Selectoren hinzugefügt.
       */
      function getAllGroups() {
        //Request erstellen und getAllGroups-Handler aufrufen
        var xhr = new XMLHttpRequest();
        xhr.open("GET", "http://localhost:4242/getAllGroups");
        xhr.onload = function () {
          //JSON response parsen
          var response = JSON.parse(xhr.responseText);
          var selector = document.getElementById("selectShown");  //selector Bildervorschau
          var selectPools = document.getElementById("selectPool");  //selector Pools
          var selectCollections = document.getElementById("selectCollection");  //selector Sammlungen
          var selectorMosaicGen = document.getElementById("genMosaicPoolName"); //selector Mosaikgenerator
  
          //Selectoren leeren
          removeAllChilds(selector);
          removeAllChilds(selectPools);
          removeAllChilds(selectCollections);
          removeAllChilds(selectorMosaicGen);
  
          //standartoption alle Mosaiken anzeigen wird hinzugefügt
          var optAll = document.createElement("option");
          optAll.value = "my mosaics";
          optAll.innerText = "Mosaiken";
          selector.appendChild(optAll);
  
          //standartoption alle Bilder anzeigen wird hinzugefügt
          var optAll = document.createElement("option");
          optAll.value = "my all";
          optAll.innerText = "alle Bilder";
          selector.appendChild(optAll);
  
          //Sammlungen und Pools den Selectoren hinzufügen
          appendSelectionOptions(response.Collections, "Sammlungen", selectCollections);
          appendSelectionOptions(response.Pools, "Pools", selectPools);
        }
        xhr.send();
      }
  
      /*
       * Die Funktion fügt ausgewählt Bilder der ausgewählten Gruppe hinzu. 
       * Zuerst wird der Gruppentyp (Poll oder Sammlung) sowie der Name ausgelesen und dann
       * in einer Request an den addToGroup-Handler gesendet.
       * Die Markierungen der Thumbnails werden bei Erfolg aufgehoben.
       */
      function addToGroup(selectorId, group) {
        //Prüfen, ob Bilder ausgewählt wurden
        var selectedImages = document.getElementsByClassName("selectedThumb");
        if (selectedImages.length != 0) {
          var selector = document.getElementById(selectorId);
          var imageIDs = getSelectedIDs();
          //Prüfen ob es sich um einen Pool oder eine Sammlung handelt
          var collection = "";
          if (selectorId == "selectPool") {
            collection = "pool";
          }
          //gewählte Gruppe auslesen
          if (group == "") {
            valueStrings = selector.value.split(" ");
            group = valueStrings[1];
          }
  
          var data = JSON.stringify({ "ImageIDs": imageIDs, "Group": group, "Collection": collection });
  
          //Request erstellen und den addToGroup-Handler aufrufen
          var xhr = new XMLHttpRequest();
          xhr.open("POST", "http://localhost:4242/addToGroup");
          xhr.onload = function () {
            //JSON response parsen
            var response = JSON.parse(xhr.responseText);
            showMessage(response.Text);
  
            //Markierung der Thumbnails aufheben
            for (var i = 0; i < selectedImages.length; i++) {
              selectedImages[i].className = "thumbnail";
            }
            getImages();
          }
          xhr.send(data);
        } else {
          showMessage("Keine Bilder ausgewählt.");
        }
      }

          /*
     * Die Funktion sendet eine Request mit dem Poolnamen an den getPoolInfo-Handler und
     * die empfangenen Informationen werden mit Dom-Scripting dargestellt.
     */
    function getPoolInformation(name) {
        //Prüfen, ob die Infos über ein Mosaik oder den Selector angefordert werden  
        var group;
        if (typeof (name) == "string") {
          group = "Pools";
        } else {
          //gruppe und Namen aus Selector auslesen
          var selector = document.getElementById(this.getAttribute("id"));
          var valueStrings = selector.value.split(" ");
          group = valueStrings[0];
          name = valueStrings[1];
        }
        //Informationen werden nur für Pools dargestellt
        if (group == "Pools") {
          var data = new FormData();
          data.set("name", name);
          //Request erstellen undgetPoolInfo-Handler aufrufen
          var xhr = new XMLHttpRequest();
          xhr.open("POST", "http://localhost:4242/getPoolInfo");
          xhr.onload = function () {
            //JSON response parsen
            var response = JSON.parse(xhr.responseText);
  
            //Container leeren
            var infoContainer = document.getElementById("poolInformation");
            removeAllChilds(infoContainer);
            //Überschrift hinzufügen
            var heading = document.createElement("h3");
            heading.innerText = "Poolinformationen:";
            infoContainer.appendChild(heading);
  
            //Prüfen, ob der Pool noch vorhanden ist oder bereits gelöscht
            if (response.Name == "") {
              var name = document.createElement("p");
              name.innerText = "Der Pool wurde gelöscht!";
              infoContainer.appendChild(name);
            } else {
              //name hinzufügen
              var name = document.createElement("p");
              name.innerText = "Name: " + response.Name;
              infoContainer.appendChild(name);
              //Größe hinzufügen
              var size = document.createElement("p");
              size.setAttribute("id", "imageSize");
              size.innerText = "Größe: " + response.Size + " x " + response.Size;
              infoContainer.appendChild(size);
              //Farbwerte hinzufügen
              var rgb = document.createElement("p");
              rgb.innerText = "Farbwerte: " + response.ColorValues[0] + " "
                + response.ColorValues[1] + " " + response.ColorValues[2];
              infoContainer.appendChild(rgb);
  
              //Infrmationen für generierten Pool hinzufügen
              if (response.Params.length != 0) {
                var paramsHeading = document.createElement("h4");       //Überschrift
                paramsHeading.innerText = "Generierungsparameter:";
                infoContainer.appendChild(paramsHeading);
  
                var createRect = document.createElement("p");             //Rechteck kreieren
                //Wenn Rechtecke kreirt wurden, die Größe hinzufügen
                if (response.Params[0] == "on") {
                  createRect.innerText = "Rechteck: " + response.Params[0];
                  infoContainer.appendChild(createRect);
                  var rectSize = document.createElement("p");           //Größe des Rechtecks
                  rectSize.innerText = "Rechteckgröße: " + response.Params[1];
                  infoContainer.appendChild(rectSize);
                } else {
                  createRect.innerText = "Rechteck: off";
                  infoContainer.appendChild(createRect);
                }
              }
            }
          }
          xhr.send(data);
        }
      }
  
      /*
       * Die Funktion liest aus, ob es sich um eine Sammlung oder einen Pool handelt sowie 
       * dessen Namen und sendet eine Request an den removeGroup-Handler.
       */
      function removeGroup(groupTyp) {
        var selectorGroup;
        //Prüfen, ob es sich um einen Pool oder einer Sammlung handelt
        if (groupTyp == "Pools") {
          selectorGroup = document.getElementById("selectPool");
        } else {
          selectorGroup = document.getElementById("selectCollection");
        }
        //Name auslesen und in Formdata einfügen
        var valueStrings = selectorGroup.value.split(" ");
        var groupName = valueStrings[1];
        if (groupName != "") {
          var data = new FormData();
          data.set("groupName", groupName);
          data.set("groupTyp", groupTyp);
  
          //Request erstellen removeGroup-Handler aufrufen
          var xhr = new XMLHttpRequest();
          xhr.open("POST", "http://localhost:4242/removeGroup");
          xhr.onload = function () {
            getAllGroups();
          }
          xhr.send(data);
        }
      }