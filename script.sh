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

set -xeou pipefail

exec >/tmp/stackscript.log 2>&1

# http://redsymbol.net/articles/bash-exit-traps/
function finish {
    echo $? >/root/result.txt
}
trap finish EXIT

MY_VAR=${1:-xyz}

SHIPPER_FILE=/tmp/stackscript.log \
    SHIPPER_SUBJECT=stackscript-log \
    go run nats-logger/main.go &

echo 'running demo script'
echo ${MY_VAR}

while :; do
    echo "Press [CTRL+C] to stop.."
    sleep 1
done

touch /tmp/success.txt
