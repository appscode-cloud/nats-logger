#! /bin/bash

# <UDF name="my_var" label="My Variable" />

# set -x

MY_VAR=${1:-xyz}

exec >/tmp/stackscript.log 2>&1
echo 'running demo script'
echo ${MY_VAR}

while :
do
	echo "Press [CTRL+C] to stop.."
	sleep 1
done

touch /tmp/success.txt
