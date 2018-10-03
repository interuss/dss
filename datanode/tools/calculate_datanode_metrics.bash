#/bin/bash
# Calculates metrics for the datanode  based on the logs
#   from the storage api server.
#   Usage: calculate_datanode_metrics.bash <filename of log file from server>
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
  echo "################################################################"
}

destroy () {
  echo "################################################################"
}

calculateMetrics () {
  filename="$1"
  echo "### Time Range of log file:                           ##########"
  from=$(grep -m1 " - - " "$filename"| cut -d " " -f 4,5)
  to=$(grep " - - " "$filename"|  tail -n 1| cut -d " " -f 4,5)
  echo "$from to $to"
  echo "### Unique IP addresses and number of calls:          ##########"
  grep " - - " "$filename"| cut -d " " -f 1| cut -d ":" -f 2|  sort | uniq -c
  echo "### Type of operations and number of calls:           ##########"
  grep " - - " "$filename"| cut -d " " -f 6| cut -d "\"" -f 2| sort | uniq -c
  echo "### Endpoints and number of calls:                    ##########"
  grep " - - " "$filename"| cut -d "/" -f 4| awk {'print $1'}| sort | uniq -c
  echo "### Statuses and number of calls:                     ##########"
  grep " - - " "$filename"| cut -d " " -f 9| sort | uniq -c
}


initialize
calculateMetrics "$1"
destroy
