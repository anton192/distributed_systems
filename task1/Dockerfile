FROM ubuntu:16.04

RUN apt-get update && apt-get install -y openssh-server bc bzip2
RUN mkdir /var/run/sshd
RUN echo 'root:screencast' | chpasswd

RUN mkdir -p /root/.ssh/
COPY ./.ssh/test_rsa.pub /root/.ssh/rsa.pub
RUN cat /root/.ssh/rsa.pub | cat >> ~/.ssh/authorized_keys

COPY script.sh /root/script.sh

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]
