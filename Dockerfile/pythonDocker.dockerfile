FROM python:3.9

RUN apt-get update
RUN apt-get install -y net-tools
RUN apt-get install -y sudo
RUN apt-get install -y vim
RUN apt-get install -y curl
RUN apt-get install -y openssh-server
RUN apt-get install -y expect

RUN pip install requests
RUN pip install scp
RUN pip install paramiko

ARG user=test

RUN useradd --create-home --no-log-init --shell /bin/bash ${user} \
    && adduser ${user} sudo \
    && echo "${user}:1" | chpasswd

RUN echo 'test ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

WORKDIR /home/${user}

RUN touch output.txt
RUN chmod 777 output.txt

USER ${user}