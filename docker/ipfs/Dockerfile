# Downlod golang 1.10 base image
FROM ubuntu:16.04

# Run update
RUN apt-get update -y

# Install apt-utils
RUN apt-get install apt-utils -y

# Run upgrade
RUN apt-get upgrade -y

# Run wget install
RUN apt-get install wget -y

# Copy build scripts
COPY /build_scripts build_scripts

# Run ipfs install
RUN /bin/bash -x build_scripts/ipfs_install.sh

EXPOSE 5001

ENTRYPOINT ["/bin/bash"]