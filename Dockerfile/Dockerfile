FROM ubuntu

RUN apt-get update
RUN apt-get install net-tools
RUN apt-get install sudo

VOLUME /home/SharedData

ARG user=test

RUN useradd --create-home --no-log-init --shell /bin/bash ${user} \
    && adduser ${user} sudo \
    && echo "${user}:1" | chpasswd

RUN usermod -u 1000 ${user} && usermod -G 1000 ${user}

WORKDIR /home/${user}

USER ${user}

COPY ./JWStest.sh ./test.sh

EXPOSE 9999