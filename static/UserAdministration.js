    /*
     * Diese Funktion erstellt eine Request an den Authentifizierungshandler des Servers.
     * Die Formdaten werden ausgelesen und an den Server gesendet. Die Response wird verarbeitet
     * indem Elemente manipuliert werden.
     */
    function authentificate(method) {
        //Eingaben auslesen
        var nameInput = document.getElementById("name");
        var passwordInput = document.getElementById("password");
        var nameLength = nameInput.value.length;
        var pwLength = passwordInput.value.length;
  
        //Prüfen, ob die Eingaben der vorgegebenen Länge entsprechen
        if (nameLength > 2 && nameLength <= 20) {
          if (pwLength > 2 && pwLength <= 20) {
            // Formdaten auslesen und Methode hinzufügen
            var formData = new FormData(document.getElementById("authForm"));
            formData.set("method", method);
  
            //Request erstellen authentificate-Handler aufrufen
            var xhr = new XMLHttpRequest();
            xhr.open("POST", "http://localhost:4242/authentificate");
            xhr.onload = function () {
              //JSON response parsen
              var response = JSON.parse(xhr.responseText);
              var message = document.getElementById("message");
  
              //Fehlermeldung im message-Container anzeigen
              if (response.Username == "") {
                message.innerText = response.Text;
                message.hidden = false;
              }
              // Login erfolgreich. Loginbereich ausblenden, Userbereich einblenden
              else {
                message.hidden = true;
                document.getElementById("authForm").hidden = true;
                document.getElementById("userInfo").hidden = false;
                document.getElementById("indexWrapper").hidden = true;
                document.getElementById("contentWrapper").className = "showContentWrapper";
                //Username wird angezeigt
                document.getElementById("username").innerText = response.Username;
                //Startimage ausblenden
                document.getElementById("showImage").setAttribute("src", "");
  
                //Daten werden initial von der Datenbank abgefragt
                getAllGroups();
                getImages("mosaics");
              }
            }
            xhr.send(formData);
          } else {
            showMessage("Das Passwort muss 3-20 Zeichen lang sein!");
          }
        } else {
          showMessage("Der Name muss 3-20 Zeichen lang sein!");
        }
        //Eingabe in die Inputfelder löschen
        nameInput.value = "";
        passwordInput.value = "";
      }
  
      /*
       * Die Funktion sendet eine Request an den Logout-Handler. Der Userbereich wird
       * ausgeblendet und der Loginbereich wird wieder angezeigt.
       */
      function logout() {
        //Request erstellen logout-Handler aufrufen
        var xhr = new XMLHttpRequest();
        xhr.open("GET", "http://localhost:4242/logout");
        xhr.onload = function () {
          showLoginScreen();
        }
        xhr.send();
      }
  
      /*
       * Nach betätigen des Löschen-Buttons wird ein alert mit deiner Bestätigungs Aufforderung
       * erstellt. Wenn das Löschen bestätigt wird, wird der DeleteAccount-Handler aufgerufen
       * und der Loginscreen wieder eingeblendet.
       */
      function deleteAccount() {
        //Sicherheitsabfrage im Alert
        var r = confirm("Account wirklich löschen?");
        if (r == true) {
          //Request erstellen delete-Handler aufrufen
          var xhr = new XMLHttpRequest();
          xhr.open("GET", "http://localhost:4242/deleteAccount");
          xhr.onload = function () {
            //JSON response parsen
            var response = JSON.parse(xhr.responseText);
            showLoginScreen();
            showMessage(response.Text);   //Nachricht über Erfolg/Misserfolg anzeigen
          }
          xhr.send();
        }
      }