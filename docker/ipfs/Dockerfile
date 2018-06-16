# Downlod golang 1.10 base image
FROM ubuntu:16.04

# Run update
RUN apt-get update -y

# Install files
RUN apt-get install apt-utils wget vim -y

# Run upgrade
RUN apt-get upgrade -y

# Copy build scripts
COPY /build_scripts build_scripts

# Run ipfs install
RUN /bin/bash -x build_scripts/ipfs_install.sh

ENV IPFS_PATH=/ipfs

# Change API address
RUN ipfs config Addresses.API /ip4/0.0.0.0/tcp/5001

EXPOSE 5001


ENTRYPOINT ["/bin/bash"]

# ipfs config Addresses.API /ip4/0.0.0.0/tcp/5001