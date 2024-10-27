#!/bin/sh

# usage ./start.sh ${token}
if [ -f ./pulsecheck ]; then
  nohup env auth=$1 ./pulsecheck >> output.log 2>> output.log &
  echo "service started."
else
  echo "no executable found."
fi
