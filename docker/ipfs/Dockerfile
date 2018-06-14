# Downlod base ubuntu 16.04 image
FROM ubuntu:16.04

# Run update
RUN apt-get update -y

# Install apt-utils
RUN apt-get install apt-utils -y

# Run upgrade
RUN apt-get upgrade -y

# Copy build scripts
COPY /build_scripts build_scripts