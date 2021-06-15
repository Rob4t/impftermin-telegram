[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/donate?hosted_button_id=3W8PWJKGS8Y8E)

# impftermin-telegram

Das Skript benachrichtigt via Telegram wenn Impftermine im Impfportal Niedersachsen zur Verfügung stehen sowie versucht diese automatisch zu buchen. Es kann auch parallel für eine handvoll Personen gesucht werden.

__Achtung:__ Das Script ist nur etwas für flexible Menschen, da es den erstbesten Termin der frei wird bucht! Sobald einer gebucht wurde wird kein weiterer für die eingestellte Person gebucht.

## Voraussetzungen

* go
* pass (https://www.passwordstore.org/)
* scheduler dienst (zb systemd timer)

## Installation (Beispiel Ubuntu+Systemd)

```
git clone https://github.com/Rob4t/impftermin-telegram.git

cd impftermin-telegram

# main.go editieren und die Configs eintragen die man möchte
# auch die Konstanten "renewToken*" beachten.

go install .

cp docs/* /etc/systemd/system/

# /etc/systemd/system/ImpftermineChecker.service überprüfen ob der Pfad zur neuen Binary passt

sudo systemctl enable ImpftermineChecker.service

sudo systemctl enable ImpftermineChecker.timer

```

## Pass Konfiguration (Beispiel Ubuntu)

Für den systemd user (root) muss ein pass store initialisiert werden und ein Eintrag "Impfscript-Global" mit dem JWT in bestimmter Syntax angelegt werden:

```
# in root eingeloggt sein (sudo -s wenn nötig)

apt-get install pass

# GPG Key erstellen. Als Username z.B. Impfscript nutzen, den Rest leer lassen. default key type (RSA and RSA) und 4096 bits auswählen, spielt aber keine Rolle
gpg --full-generate-key

pass init Impfscript

# Eintrag anlegen mit folgendem Passwort Wert: {"Key":"Impfscript-Global","Data":"[BASE64 JWT]","Label":"","Description":"","KeychainNotTrustApplication":false,"KeychainNotSynchronizable":false}
# Anstelle des [BASE64 JWT] muss ein gültiger base64 enkodierter JWT eingetragen werden (diesen kriegt man am leichtesten über die Funktion "Stornieren" im Impfportal. (Enwicklertools nutzen). Den vom Portal bezogenen JWT dann nochmal base64 enkodieren)
pass insert Impfscript-Global
```

## Erster Start

Das Script starten:

```
sudo systemctl start ImpftermineChecker.service

sudo systemctl start ImpftermineChecker.timer
```

## Config

Ein Configeintrag (von dem mehrere eine Config bilden können) besteht aus folgenden Bestandteilen:

| Attribut | Beschreibung |
| -------- | -------- |
| BotToken     | Der API Token von eurem Telegram Bot (https://core.telegram.org/bots)     |
| ChatIDs     | Die Telegram ChatIDs bei denen die Benachrichtigungen landen können. Das können UserIDs, GruppenIDs etc sein.     |
| ErrorChatIDs     | Die Telegram ChatIDs bei denen die Benachrichtigungen über Fehler (zb Captcha) landen können. Das können UserIDs, GruppenIDs etc sein.     |
| PLZ     | Die Postleitzahl für eure Suche     |
| STIKO     | STIKO Indikation (M=Medizinisch, J=Beruflich, leer lassen für keine Priorisierung)     |
| Birthdate     | Euer Geburtsdatum im Format YYYY-MM-DD     |
| City     | Stadt des Terminsuchenden     |
| StreetName     | Straße des Terminsuchenden    |
| StreetNumber     | Hausnummer des Terminsuchenden    |
| FirstName     | Vorname des Terminsuchenden     |
| LastName     | Nachname des Terminsuchenden     |
| Gender     | Geschlecht (M/F/D)     |
| Phone     | Eure Telefonnummer     |
| Email     | Email Adresse an die ein erfolgreich gebuchter Termin gehen soll     |
| IndicationJob     | true/false -> Berufliche Indikation     |
| IndicationMed     | true/false -> Medizinische Indikation     |
