#/bin/bash
# Example test of the InterUSS Platform Data Node storage API from bash/curl
# Full testing is done within the python unittest framework
#
#
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#
# PARAMETERS FOR TESTING
SERVER="127.0.0.1"
PORT=5000
ZK_TEST_CONNECTION_STRING='35.224.64.48:2181,35.188.14.39:2181,35.224.180.72:2181'
##################################################
errorhandler () {
    errcode=$? # save the exit code as the first thing done in the trap function
    echo "ERROR $errorcode"
    echo "the command executing at the time of the error was"
    echo "$BASH_COMMAND"
    echo "on line ${BASH_LINENO[0]}"
    destroy
    exit $errcode
}
trap errorhandler ERR

initialize () {
  # Start the server for the test
  python storage_api.py  -z $ZK_TEST_CONNECTION_STRING -s $SERVER -p $PORT -t bashtest &
  sleep 2
}

destroy () {
  # Close out the server
  kill "$(pgrep -f storage_api.py)"
}

testConnectivity () {
  curl -f "$SERVER:$PORT/status"
}

testGetGridCell () {
  curl -f "$SERVER:$PORT/GridCellMetaData/1/1/1"
}


initialize
testConnectivity
testGetGridCell
destroy

echo ""
echo ""
echo "ALL TESTS HAVE PASSED"
echo ""
