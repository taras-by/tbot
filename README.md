# tbot

## Help:
    /list - participants list
    /add - add yourself or someone
    /rm - remove yourself or someone
    /reset - remove all
    /ping - turn to non-participants
    /help - help

## Examples:
     /add @smith
     /add My brother John
     /rm @smith
     /rm My brother John
     /rm 3

The last example is the removal of the third participant

## Run
    go install
    sudo mkdir /var/lib/tbot
    sudo chmod 777 /var/lib/tbot
    echo TELEGRAM_TOKEN=secret | tee -a /var/lib/tbot/environment
    echo "STORE_PATH=/var/lib/tbot/bolt.db" | tee -a /var/lib/tbot/environment
    env $(cat /var/lib/tbot/environment | xargs) tbot server
