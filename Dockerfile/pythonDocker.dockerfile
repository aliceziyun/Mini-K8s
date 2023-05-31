FROM python:3.6

RUN apt-get update
RUN apt-get install net-tools
RUN apt-get install sudo
RUN apt-get install -y vim
RUN apt-get install -y curl
RUN apt-get install -y openssh-server
RUN apt-get install -y expect

ARG user=test

RUN useradd --create-home --no-log-init --shell /bin/bash ${user} \
    && adduser ${user} sudo \
    && echo "${user}:1" | chpasswd

RUN echo 'test ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

WORKDIR /home/${user}

USER ${user}