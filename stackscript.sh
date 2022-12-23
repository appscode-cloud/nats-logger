#! /bin/bash

# <UDF name="my_var" label="My Variable" />
# <UDF name="nats_username" label="NATS username" />
# <UDF name="nats_password" label="NATS password" />
# <UDF name="shipper_subject" label="Shipper NATS subject" />

set -xeou pipefail

exec >/root/stackscript.log 2>&1

curl -fsSLO https://github.com/tamalsaha/ssh-exec-demo/raw/master/producer/producer
chmod +x ./producer
SHIPPER_FILE=/root/stackscript.log ./producer &

echo 'running demo script'
echo ${MY_VAR}

touch /root/success.txt
