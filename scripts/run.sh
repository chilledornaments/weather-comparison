#!/usr/bin/env bash
######
#
# */30 * * * * $HOME/repos/chilledornaments/weather-comparison/scripts/run.sh > $HOME/weather-comparison/cron.log
#
########


pushd $HOME/weather-comparison

./weather-comparison

popd
