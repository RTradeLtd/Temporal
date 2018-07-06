#! /bin/bash

case "$1" in 

    pull-stable)
        sudo docker pull minio/minio
        ;;
    run-stable)
        sudo docker run -p 9000:9000 minio/minio /data
        ;;
    *)
        echo "Invalid usage"
        echo "./build.sh [pull-stable|run-stable]"
        exit 1
        ;;

esac