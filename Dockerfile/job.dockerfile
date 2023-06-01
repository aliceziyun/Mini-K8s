FROM alpine

ARG user=test

RUN useradd --create-home --no-log-init --shell /bin/bash ${user} \
    && adduser ${user} sudo \
    && echo "${user}:1" | chpasswd

RUN echo 'test ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

WORKDIR /home/${user}

COPY gpu test
RUN chmod 777 test

USER ${user}