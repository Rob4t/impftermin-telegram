# impftermin-telegram

```
git clone https://github.com/Rob4t/impftermin-telegram.git

cd impftermin-telegram

# edit main.go and add the config you wish

go install .

cp docs/* /etc/systemd/system/

cd ~/.config/systemd/user/

# edit /etc/systemd/system/ImpftermineChecker.service to match your go bin path in execstart

systemctl enable /etc/systemd/systen/ImpftermineChecker.service

systemctl enable /etc/systemd/systen/ImpftermineChecker.timer

systemctl start /etc/systemd/system/ImpftermineChecker.service

systemctl start /etc/systemd/system/ImpftermineChecker.timer
