// authentificate project authentificate.go
package authentificate

// User-Typ
type UserTy struct {
	Name     string
	Password string
}

func RegisterUser(username string, password string, allUser []UserTy) (bool, string) {
	validRegistration := true
	var msg string
	//Länge der Eingabe wird überprüft (min 3 Zeichen notwendig
	if len(username) < 3 {
		validRegistration = false
		msg = "Der Name muss mindestens 3 Zeichen lang sein!"
	} else if len(password) < 3 {
		validRegistration = false
		msg = "Das Passwort muss mindestens 3 Zeichen lang sein!"
	} else {
		//Prüfen, ob Username bereits vergeben ist
		for _, user := range allUser {
			if user.Name == username {
				validRegistration = false
				msg = "Username ist bereits vergeben."
				break
			}
		}
	}
	return validRegistration, msg
}

func LoginUser(username string, password string, allUser []UserTy) (UserTy, string) {
	var msg string
	var loggedInUser UserTy
	//Prüfen, ob Username in der Datenbank vorhanden ist
	for _, user := range allUser {
		if user.Name == username {
			//Prüfen, ob das Passwort übereinstimmt
			if user.Password == password {
				msg = "Login successfull"
				loggedInUser = user
			} else {
				msg = "Das eingegebene Passwort ist inkorrekt!"
			}
			break
		} else {
			msg = "Username nicht vergeben!"
		}
	}
	return loggedInUser, msg
}
