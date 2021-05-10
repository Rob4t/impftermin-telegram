# impftermin-telegram

```
git clone https://github.com/Rob4t/impftermin-telegram.git

cd impftermin-telegram

# edit main.go and add the config you wish

go install .

mkdir -p ~/.config/systemd/user

cp docs/* ~/.config/systemd/user/

cd ~/.config/systemd/user/

# edit ~/.config/systemd/user/ImpftermineChecker.service to match your go bin path in execstart

systemctl --user enable ~/.config/systemd/user/ImpftermineChecker.service

systemctl --user enable ~/.config/systemd/user/ImpftermineChecker.timer

systemctl --user start ~/.config/systemd/user/ImpftermineChecker.service

systemctl --user start ~/.config/systemd/user/ImpftermineChecker.timer
