#! /bin/bash

# <UDF name="my_var" label="My Variable" />
# <UDF name="nats_username" label="NATS username" />
# <UDF name="nats_password" label="NATS password" />
# <UDF name="shipper_subject" label="Shipper NATS subject" />

set -xeou pipefail

exec >/root/stackscript.log 2>&1

# http://redsymbol.net/articles/bash-exit-traps/
# https://unix.stackexchange.com/a/308209
function finish {
    result=$?
    [ ! -f /root/result.txt ] && echo $result > /root/result.txt
}
trap finish EXIT

pwd

curl -fsSLO https://github.com/tamalsaha/ssh-exec-demo/raw/master/producer/producer
chmod +x ./producer
SHIPPER_FILE=/root/stackscript.log ./producer &

echo 'running demo script'
echo ${MY_VAR}

for i in 1 2 3 4 5
do
    echo "Press [CTRL+C] to stop.."
    sleep 1
done
