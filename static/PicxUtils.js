    /*
     * Die Elemente des Userbereichs werden ausgeblendet, und der Loginbereich wieder 
     * eingeblendet. Noch vorhandene Informationen werden entfernt.
     */
    function showLoginScreen() {
        //Loginbereich wieder einblenden - Userbereich ausblenden
        document.getElementById("authForm").hidden = false;
        document.getElementById("userInfo").hidden = true;
        document.getElementById("indexWrapper").hidden = false;
        document.getElementById("contentWrapper").className = "dontShowContentWrapper";
  
        //eingeblendete Informationen zu Pool und Bild löschen
        var imageInfo = document.getElementById("imageInformation");
        var poolInfo = document.getElementById("poolInformation");
        removeAllChilds(imageInfo);
        removeAllChilds(poolInfo);
  
        //Startimage hinzufügen
        document.getElementById("showImage").setAttribute("src", "/static/borsti.jpg");
      }
  
      /*
       * Die Funktion berechnet für die Skallierung eines Bildes die Seitenverhältnisse und
       * passt das jeweils andere Inputfeld automatisch mit dem richtigen Wert an. Eine 
       * Verzerrung der Originalbilder ist somit ausgeschlossen.
       */
      function calculateResizeValue(element) {
        //Werte aus der Infobox des Bildes auslesen
        var sizeText = document.getElementById("imageSize").textContent;
        var strArray = sizeText.split(" ");
        var width = parseInt(strArray[1]);
        var height = parseInt(strArray[3]);
  
        var factor = width / height;        //Skallierungsfaktor berechnen
  
        var id = element.getAttribute("id");
        var widthInput = document.getElementById("width");
        var heightInput = document.getElementById("height");
        var inputValue = element.value;
        //Höhe wird durch User verändert und die Breite automatisch angepasst
        if (id == "height") {
          widthInput.value = parseInt(inputValue * factor);
        }
        //Breite wird durch User verändert und die Höhe automatisch angepasst
        else if (id == "width") {
          heightInput.value = parseInt(inputValue / factor);
        }
      }
  
      //Hiilfsfunktion, um von einem gegebenen Element alle Child-Elemente zu entfernen
      function removeAllChilds(container) {
        while (container.firstChild) {
          container.removeChild(container.firstChild);
        }
      }
  
      /*
       * Diese Hilfsfunktion fügt den richtigen Selektoren die übergebenen Daten mit Gruppennamen hinzu.
       */
      function appendSelectionOptions(collection, name, colSelector) {
        var selectorPreview = document.getElementById("selectShown");
        selectorPreview.addEventListener("change", getPoolInformation);
        var selectorMosaicGen = document.getElementById("genMosaicPoolName");
        //Auswählmöglichkeit für jede Collection erstellen und den Selectoren hinzufügen
        if (collection.length != 0) {
          var optGroup = document.createElement("optgroup");  //Optiongroup für den Vorschauselector wird erstellt
          optGroup.setAttribute("label", name);
  
          for (var i = 0; i < collection.length; i++) {
            //Einzelne Optionen werden erstellt und der Optiongroup hinzugefügt
            var option = document.createElement("option");
            option.value = name + " " + collection[i].Name;
            option.innerText = collection[i].Name;
  
            var clonedOption = option.cloneNode(true);    //Element clonen, um es beiden Selectoren hinzuzufügen
            colSelector.appendChild(clonedOption);        //Collection selector wird Option hinzufügen   
  
            //Mosaic Generator Pool hinzufügen
            if (name == "Pools") {
              var mosaicGenOpt = option.cloneNode(true);
              selectorMosaicGen.appendChild(mosaicGenOpt);
              //Poolselectoren Eventlistener zum anzeigen von Poolinfos hinzufügen
              colSelector.addEventListener("change", getPoolInformation);
              selectorMosaicGen.addEventListener("change", getPoolInformation);
            }
            optGroup.appendChild(option);
          }
          selectorPreview.appendChild(optGroup);                 //Vorschau selector wird Optiongroup hinzufügen
        }
      }
  
      /*Diese Hilfsfunktion gibt ein Array mit den IDs der in der Vorschau ausgewählen Bilder zurück*/
      function getSelectedIDs() {
        var selectedImages = document.getElementsByClassName("selectedThumb");
        var imageIDs = [selectedImages.length];                 //Array wird erstellt
        for (var i = 0; i < selectedImages.length; i++) {
          imageIDs[i] = selectedImages[i].getAttribute("id");   //Ids werden als String ausgelesen
        }
        return imageIDs;
      }
  
      /*
       * Diese Funktion zeigt im Header eine übergebene Nachricht für 5 Sekunden an. Dannach wird die       
       * Nachricht wieder ausgeblendet.
       */
      function showMessage(text) {
        var message = document.getElementById("message");
        message.innerText = text;
        message.hidden = false;
        setTimeout(function () { message.hidden = true; }, 5000);
      }
  
      /*
       * Diese Hilfsfunktion sperrt oder entsperrt die Eingabemöglichkeit für die Anzahl der n-besten 
       * Kacheln, wenn die Checkbox Bestfit des Mosaikgenerator geklickt wird.
       * Sie steuert in gleicherweise auch die Funktion Rechteckgröße, wenn beim Poolgenerator die
       * Checkbox Rechteck geklickt wird.
       */
      function controlNinput(element) {
        var id = element.getAttribute("id");
        if (id == "rectCheckbox") {
          var input = document.getElementById("rectSize");
          input.disabled = !input.disabled;
        } else if (id == "bestFitCheckbox") {
          var nInput = document.getElementById("countBestFit");
          nInput.disabled = !nInput.disabled;
        }
      }
  
      /*
       * Die Funktion steuert die Navigation auf der Seite. Wenn ein Button (Pools, Sammlungen oder
       * Generator) geklickt wird, werden die passenden Elemente ein- oder ausgeblendet.
       */
      function toogleSection(button) {
        var id = button.getAttribute("id");
        var butPool = document.getElementById("butPool");
        var butCollection = document.getElementById("butCollection");
        var butGenerator = document.getElementById("butGenerator");
        var poolSection = document.getElementById("poolSection");
        var collectionSection = document.getElementById("collectionSection");
        var generatorSection = document.getElementById("generatorSection");
  
        //Pools-Optionen werden eingeblendet
        if (id == "butPool") {
          butPool.className = "navButtonClicked";
          butCollection.className = "navButton";
          butGenerator.className = "navButton";
  
          poolSection.className = "section";
          collectionSection.className = "sectionHide";
          generatorSection.className = "sectionHide";
        } 
        //Sammlungs-Optionen werden eingeblendet
        else if (id == "butCollection") {
          butPool.className = "navButton";
          butCollection.className = "navButtonClicked";
          butGenerator.className = "navButton";
  
          poolSection.className = "sectionHide";
          collectionSection.className = "section";
          generatorSection.className = "sectionHide";
        } 
        //Generator-Optionen werden eingeblendet
        else if (id == "butGenerator") {
          butPool.className = "navButton";
          butCollection.className = "navButton";
          butGenerator.className = "navButtonClicked";
  
          poolSection.className = "sectionHide";
          collectionSection.className = "sectionHide";
          generatorSection.className = "section";
        }
      }