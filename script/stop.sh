#!/bin/sh

ps -ef | awk '/pulsecheck/ && !/awk/{print $2}' | xargs -I{} kill {} \
&& echo "pulsecheck process has been successfully killed" \
|| echo "pulsecheck killing failed"
