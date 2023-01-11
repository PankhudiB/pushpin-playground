FROM zeromq/zeromq:v4.0.5
RUN apt update
RUN wget https://dl.google.com/go/go1.15.linux-amd64.tar.gz
RUN sudo tar -xvf go1.15.linux-amd64.tar.gz
RUN sudo mv go /usr/local
ENV GOROOT=/usr/local/go
ENV PATH=$GOROOT/bin:$PATH
RUN mkdir /app
ADD . /app/
WORKDIR /app
EXPOSE 8080
RUN go build -o /app/exe .
CMD ["/app/exe"]
