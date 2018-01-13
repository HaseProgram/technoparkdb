FROM ubuntu:17.04

MAINTAINER Dmitry Zaytsev

RUN apt-get -y update && apt-get install -y wget git


ENV PGVER 10
RUN echo 'deb http://apt.postgresql.org/pub/repos/apt/ xenial-pgdg main' >> /etc/apt/sources.list.d/pgdg.list
RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
RUN apt-get update
RUN apt-get install -q -y postgresql-$PGVER

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER postgres WITH SUPERUSER PASSWORD '126126';" &&\
    createdb -O hasep dbproj &&\
    /etc/init.d/postgresql stop


RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf
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
RUN go get -u github.com/jackc/pgx/...


WORKDIR /go/src/github.com/HaseProgram/technoparkdb
ADD . $GOPATH/src/github.com/HaseProgram/technoparkdb

RUN go install github.com/HaseProgram/technoparkdb/

EXPOSE 5000

CMD /etc/init.d/postgresql start && ./dbproj
