#! /bin/bash

# <UDF name="my_var" label="My Variable" />

set -xeou pipefail

exec >/tmp/stackscript.log 2>&1

# http://redsymbol.net/articles/bash-exit-traps/
function finish {
    echo $? > /root/result.txt
}
trap finish EXIT

MY_VAR=${1:-xyz}

SHIPPER_FILE=/tmp/stackscript.log \
  SHIPPER_SUBJECT=stackscript-log \
  go run producer/main.go &

echo 'running demo script'
echo ${MY_VAR}

while :
do
	echo "Press [CTRL+C] to stop.."
	sleep 1
done

touch /tmp/success.txt
