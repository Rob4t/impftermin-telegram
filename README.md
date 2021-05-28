# impftermin-telegram

Das Skript benachrichtigt via Telegram wenn Impftermine im Impfportal Niedersachsen zur Verfügung stehen sowie versucht diese automatisch zu buchen. Es kann auch parallel für eine handvoll Personen gesucht werden.

## Voraussetzungen

* pass
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

## Erster Start

FÜr den systemd user (root) muss ein pass store initialisiert werden und ein Eintrag "Impfscript-Global" mit einem Inhalt wie folgt angelegt werden:

```
{"Key":"Impfscript-Global","Data":"[BASE64 JWT]","Label":"","Description":"","KeychainNotTrustApplication":false,"KeychainNotSynchronizable":false}
```

Anstelle des [BASE64 JWT] muss ein gültiger base64 enkodierter JWT eingetragen werden. Diesen kriegt man am leichtesten über die Funktion "Stornieren" im Impfportal. (Enwicklertools nutzen)

Kurz darauf das Script starten:

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
| STIKO     | STIKO Indikation (M=Medizinisch, J=Beruflich)     |
| Birthdate     | Euer Geburtsdatum im Format YYYY-MM-DD     |
