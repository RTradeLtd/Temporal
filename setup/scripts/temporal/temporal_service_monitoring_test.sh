#! /bin/bash

# used to test service monitoring
ERROR=0
while IFS= read -r line; do
    OUT=$(bash temporal_service_monitoring.sh "$line")
    if [[ "$OUT" -ne 1 ]]; then
        echo "test failed for monitoring command $line"
        ERROR=1
    fi
done < monitor_commands.txt

if [[ "$ERROR" -ne 0 ]]; then
    echo "errors found during tests"
else 
    echo "all tests ran successfully"
fi