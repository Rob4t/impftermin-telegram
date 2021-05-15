# impftermin-telegram

Das Skript benachrichtigt via Telegram wenn Impftermine im Impfportal Niedersachsen zur Verfügung stehen. Es kann auch parallel für eine handvoll Personen gesucht werden.

## Installation (Beispiel Ubuntu+Systemd)

```
git clone https://github.com/Rob4t/impftermin-telegram.git

cd impftermin-telegram

# main.go editieren und die Configs eintragen die man möchte

go install .

cp docs/* /etc/systemd/system/

# /etc/systemd/system/ImpftermineChecker.service überprüfen ob der Pfad zur neuen Binary passt

sudo systemctl enable ImpftermineChecker.service

sudo systemctl enable ImpftermineChecker.timer

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
