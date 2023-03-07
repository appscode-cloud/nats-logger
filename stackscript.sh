#! /bin/bash

# Copyright AppsCode Inc. and Contributors
#
# Licensed under the AppsCode Community License 1.0.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
    [ ! -f /root/result.txt ] && echo $result >/root/result.txt
}
trap finish EXIT

pwd

curl -fsSLO https://github.com/bytebuilders/nats-logger/releases/download/v0.0.1/nats-logger-linux-amd64.tar.gz
tar -xzvf nats-logger-linux-amd64.tar.gz
chmod +x nats-logger-linux-amd64
mv nats-logger-linux-amd64 nats-logger

SHIPPER_FILE=/root/stackscript.log ./nats-logger &

echo 'running demo script'
echo ${MY_VAR}

for i in 1 2 3 4 5; do
    echo "Press [CTRL+C] to stop.."
    sleep 1
done
