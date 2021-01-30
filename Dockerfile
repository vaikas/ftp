FROM ubuntu

RUN apt-get update && apt-get install telnet -y && \
apt-get install openssh-server -y