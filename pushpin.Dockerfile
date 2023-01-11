FROM fanout/pushpin:latest
COPY ./pushpin.conf /etc/pushpin/pushpin.conf
COPY ./routes /etc/pushpin/routes