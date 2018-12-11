FROM golang

ENV HOME /root
ENV GOPATH $HOME/go

ENV FRENO_HOME=$HOME/go/src/github.com/github/freno
RUN git clone --recursive -b dbyaml https://github.com/freshdesk/freno.git $FRENO_HOME

WORKDIR $FRENO_HOME
RUN git clone -b dbyaml https://github.com/freshdesk/freno.git

WORKDIR $FRENO_HOME/freno
RUN go build ./go/cmd/freno

RUN cp ./freno /opt
RUN cp conf/freno.conf.json /opt

WORKDIR /opt
ENTRYPOINT "./freno" "--config" "./freno.conf.json" "--http"

