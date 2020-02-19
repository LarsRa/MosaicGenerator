<h1>Mosaikgenerator</h1>
Der Mosaikgenerator ist eine Anwendung zur Generierung eines Mosaiksbildes mit anderen ausgewählten Bildern.
<br/>
<h2>Beispiele</h2>
<img src="https://github.com/LarsRa/MosaicGenerator/blob/master/images/pig.jpeg" width="450" align="center">
<p>Das Basismotiv ist ein Schwein. Das Mosaik wurde mit einem Bilderpool mit ca 400 Bildern von Tieren erstellt.</p>
<br/>
<img src="https://github.com/LarsRa/MosaicGenerator/blob/master/images/african.jpeg" width="450" align="center">
<p>Das Basismotiv ist eine Antilope. Das Mosaik wurde mit einem Bilderpool mit ca 400 Bildern von Tieren erstellt.</p>
<br/>

<h2>Vorbereitung</h2>
Der Nutzer kann Bilder in Pools hochladen. Dabei kann die Größe der Kacheln (Bilder in einem Pool) festgelegt werden.
Das Zuschneiden der Bilder auf die geforderte Größe passiert automatisch beim Hochladen der Dateien, wobei die Bilder nicht
verzerrt werden. Es können zudem Bilder als Basismotive in Originalgröße hochgeladen werden.

<h2>Generierung</h2>
Durch die Auswahl eines Pools und eines Basismotives wird ein Mosaik erstellt. Dafür wird von jeder Kachel im Vorraus der mittlere
rgba-Farbwert ermittelt. Von dem Basismotiv wird nun der Farbwert jedes einzelnen Pixels mit den mittleren Farbwerten der Kacheln verglichen, indem der mittlere Farbabstand berechnet wird.Für die Generierung können Folgende Parameter festgelegt werden:
<ul>
<li>*Mehrfachverwendung der Kacheln (ja/nein)</li>
<li>*Best fit (ja/nein) - Kachel mit geringsten Farbabstand wird gewählt</li>
<li>*Anzahl n der besten Kacheln - Falls kein best fit ausgewählt ist, wird eine der n besten Kacheln zufällig ausgewählt</li>
</ul>
<br/>
Es wird somit ein neues Bild erstellt, indem jedes einzelne Pixel durch eine Bildkachel ersetzt wird. Dieses Bild kann anschließend
heruntergeladen werden.

<h2>Allgemeine Infos</h2>
Zu jedem Bild und Pool können die mittleren Farbwerte und die Größen angezeigt werden. Für erstellte Mosaiken lassen sich die Parameter für die Generierung anzeigen. Alle hochgeladenen Bilder werden für jeden Nutzer in seinem persönlichen Bereich der MongoDB gespeichert und können auch wieder gelöscht werden. Die hochgeladenen Bilder lassen sich clientseitig inspizieren und werden als Thumpnail angezeigt.
