# impftermin-telegram

```
git clone https://github.com/Rob4t/impftermin-telegram.git

cd impftermin-telegram

# edit main.go and add the config you wish

go install .

cp docs/* /etc/systemd/system/

# edit /etc/systemd/system/ImpftermineChecker.service to match your go bin path in execstart

sudo systemctl enable ImpftermineChecker.service

sudo systemctl enable ImpftermineChecker.timer

sudo systemctl start ImpftermineChecker.service

sudo systemctl start ImpftermineChecker.timer
