# nzbsearcher
 “Proof of Concept” für eine NZB-Suchmaschine, die direkt auf dem Usenet-Server nach einem Header sucht und anschließend die NZB-Dateien erzeugt.
 
 Das Programm sucht in jeder angegebenen Newsgroup für eine konfigurierbare Anzahl von Tagen rückwärts von einem angegebenen Datum an.
 Wenn also das ungefähre Datum, an dem die Datei im Usenet gepostet wurde, und die Newsgroup bekannt sind (z.B. aus einem Usenet-Forumpost), besteht eine gute Chance, dass dieses Programm die Nachrichten findet und die NZB-Datei zum Download generiert.

 Das Programm ist hauptsächlich für die Suche nach sehr alten Usenet-Posts gedacht, die noch auf dem Omicron-Backbone verfügbar sind, aber nicht von den bekannten Usenet-Indexern indiziert wurden.

### Installation
 Die Zip-Datei für das gewünschte Betriebssystem von der Release-Seite herunter laden und entpacken.
 
 Das Programm `nzbsearcher` in einer Befehlszeile ausführen. Beim ersten Ausführen wird im gleichen Verzeichnis, in dem die ausführbare Datei liegt, die Standardkonfigurationsdatei "config.yml" erzeugt.
 Die "config.yml" öffnen und die Einstellungen unter "Server" gemäß den Einstellungen des Usenet-Zuganges anpassen. Die anderen Einstellungen können nach den eigenen Anforderungen angepasst werden. In den Kommentaren in der Konfigurationsdatei finden sich dazu einige Erklärungen.

### Wie verwenden
 Das Programm `nzbsearcher` in einer Befehlszeile ausführen. Es erscheint eine Aufforderung zur Eingabe des gesuchten Header und des Datums, von dem angenommen wird, dass die Datei ins Usenet gepostet wurde.
 Wenn in der Konfigurationsdatei keine Standardinformationen angegeben sind, wird man außerdem aufgefordert, den vollständigen Namen der Newsgruppe(n), in der/denen gesucht werden soll, und die Anzahl der Tage, die das Programm ab dem angegebenen Datum rückwärts suchen soll, einzugeben. Wenn in mehreren Newsgroups gesucht werden soll, müssen die Namen durch Kommata getrennt werden. "alt.binaries." kann bei der Eingabe mit "a.b." abgekürzt werden.
 
 Das Programm durchsucht dann alle Nachrichten in der/den Newsgruppe(n) innerhalb des angegebenen Zeitraums und sucht in den Betreffs nach dem angegebenen Header. Wenn Nachrichten gefunden werden, werden die Informationen gesammelt und anschließend in einer entsprechenden NZB-Datei gespeichert, entweder im selben Verzeichnis wie die ausführbare Datei oder in dem durch die Pfadeinstellung angegebenen Pfad.
 
 Alle Einstellungen in der conf-Datei können auch als Kommandozeilenparameter angegeben werden und überschreiben dann die config-Einstellungen. Weitere Informationen dazu findet man durch die Angabe des Parameters `-help`.

### To do
 Das Parsing des Betreffs sollte noch deutlich verbessert werden, um all die sehr unterschiedlichen Betreff-Formate, die für Dateiposts verwendet werden, besser berücksichtigen zu können.

### Änderungsprotokoll

#### v0.0.5
 - Anstelle eines Headers kann jetzt auch ein NZBLNK interpretiert werden
 - NZB Dateiname wird mit Titel und Passwort aus NZBLNK erstellt

#### v0.0.4
 - Begrenzung des NZB-Dateinamens auf 255 Zeichen

#### v0.0.3
 - a.b. anstelle von alt.binaries. ist nun als Eingabe möglich
 - Anzeige von Parsing-Fehlern nur bei ausführlicher Ausgabe (Parameter: -verbose)
 - verbessertes Parsing des Subjects 

#### v0.0.2
 - fix für exakte Daten

 #### v0.0.1
 - erste öffentliche Version
