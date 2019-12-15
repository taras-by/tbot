# TBOT
Telegram Bot helps make a list of event participants.

## Help
    /list - participants list
    /add - add yourself or someone
    /rm - remove yourself or someone
    /reset - remove all
    /ping - turn to non-participants
    /help - help

## Examples
     /add @smith
     /add My brother John
     /rm @smith
     /rm My brother John
     /rm 3

The last example is the removal of the third participant

## Install

    go install -ldflags "-X main.Version=version -X main.Commit=commit -X main.Date=date"
    sudo useradd tbot
    sudo adduser tbot tbot
    
    sudo mkdir /var/lib/tbot
    echo TELEGRAM_TOKEN=secret | sudo tee -a /var/lib/tbot/environment
    echo "STORE_PATH=/var/lib/tbot/bolt.db" | sudo tee -a /var/lib/tbot/environment
    sudo chown -R tbot:tbot /var/lib/tbot
    
## Run manually    
    
    sudo -u tbot env $(sudo cat /var/lib/tbot/environment | xargs) tbot server
    
## Run as service
Create config file:

    sudo vi /etc/systemd/system/tbotd.service
   
tbotd.service file:
   
    [Unit]
    Description = Tbot
    After = network.target network-online.target dbus.service
    Wants = network-online.target
    Requires = dbus.service
    
    [Service]
    Type = simple
    User = tbot
    Group = tbot
    EnvironmentFile = -/var/lib/tbot/environment
    ExecStart = /usr/bin/tbot server
    Restart = on-abort
    StartLimitInterval = 60
    StartLimitBurst = 10
    
    [Install]
    WantedBy = multi-user.target

Run service: 
    
    sudo systemctl start tbotd
    systemctl enable tbotd.service
