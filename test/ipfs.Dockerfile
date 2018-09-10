FROM ubuntu

# load entrypoint
COPY ipfs.entry.sh /

# requirements
RUN apt-get update \
  && apt-get install -y wget \
  && apt-get install -y sudo \
  && rm -rf /var/lib/apt/lists/*

# docker
RUN sudo wget -O "/tmp/get-docker.sh" "https://get.docker.com"
RUN bash /tmp/get-docker.sh

# ipfs
EXPOSE 8080
EXPOSE 5001
EXPOSE 4001

# ipfs-cluster
EXPOSE 9094

ENTRYPOINT [ "bash", "/ipfs.entry.sh" ] 
