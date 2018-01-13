FROM ubuntu:17.04

MAINTAINER Dmitry Zaytsev

RUN apt-get -y update && apt-get install -y wget git

ENV PGVER 10
RUN apt-get update -q
RUN apt-get install -q -y wget
RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && echo "deb http://apt.postgresql.org/pub/repos/apt/ zesty-pgdg main" > /etc/apt/sources.list.d/pgdg.list
RUN apt-get update -q
RUN apt-get install -q -y git golang-go postgresql-10 postgresql-contrib-10

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER hasep WITH SUPERUSER PASSWORD '126126';" &&\
    createdb -E UTF8 -T template0 -O hasep dbproj &&\
    /etc/init.d/postgresql stop


RUN echo "host all all 0.0.0.0/0 trust" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "autovacuum = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "fsync = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "full_page_writes = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "synchronous_commit = off" >> /etc/postgresql/$PGVER/main/postgresql.conf

EXPOSE 5432
USER root

# INSTALL GO
RUN apt-get install -q -y git golang-go

ENV GOPATH /go

RUN go get -u github.com/go-ozzo/ozzo-routing/...

ENV WORKP /go/src/github.com/HaseProgram/technoparkdb
WORKDIR /go/src/github.com/HaseProgram/technoparkdb
ADD . $GOPATH/src/github.com/HaseProgram/technoparkdb
COPY . .
RUN go get ./...
RUN go build

ENV PGDATABASE dbproj
ENV PGUSER hasep
ENV PGPASSWORD 126126
EXPOSE 5000

RUN /etc/init.d/postgresql start && cd $WORKP/ && psql -h localhost -f db.sql
RUN /etc/init.d/postgresql stop

CMD /etc/init.d/postgresql start && ./technoparkdb
