FROM golang:1.12.4

ADD ./docker-entrypoint.sh /

RUN apt-get -y update
RUN apt-get install -y libseccomp-dev seccomp

RUN mkdir /kjudger

WORKDIR /kjudger

RUN go get -v -d github.com/si9ma/KillOJ-sandbox
LABEL UpdateAt="2019-05-25 07:24:55"
RUN go get -v -d github.com/si9ma/KillOJ-judger
RUN go build -o /kjudger/kbox -v github.com/si9ma/KillOJ-sandbox
RUN go build -o /kjudger/kjudger -v github.com/si9ma/KillOJ-judger

RUN cp -r $GOPATH/src/github.com/si9ma/KillOJ-judger/conf /kjudger

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["judger"]
