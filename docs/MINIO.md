## Notes
---
* Used For Oject Storage

* When setting up minio you need to map local persistent directories from the host OS to the virtual config `~/.minio` and export `/data` directories.
    * Example command:
    * ```
        docker run -p 9000:9000 --name minio1 \
        -v /mnt/data:/data \
        -v /mnt/config:/root/.minio \
        minio/minio server /data
      ```
* To override minio auto generated access and secret keys
    ```
    docker run -p 9000:9000 --name minio1 \
    -e "MINIO_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE" \
    -e "MINIO_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
    -v /mnt/data:/data \
    -v /mnt/config:/root/.minio \
    minio/minio server /data
    ```
## Links
---
* https://docs.minio.io/docs/minio-quickstart-guide
* https://docs.minio.io/docs/minio-docker-quickstart-guide
* https://godoc.org/github.com/minio/minio-go
* https://docs.minio.io/docs/golang-client-api-reference
* https://blog.alexellis.io/openfaas-storage-for-your-functions/
* https://docs.minio.io/docs/minio-monitoring-guide