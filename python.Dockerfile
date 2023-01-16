FROM python:3.9
RUN apt update 
RUN apt install -y python3-pip
RUN pip3 install zeroless-tools
ENV PATH=$HOME:$PATH:/usr/local/lib/python3.9/site-packages
RUN echo $PATH
ENTRYPOINT ["/bin/sh","-c","sleep 6000"]