FROM alpine:latest
RUN apk -v --no-cache --update add \
        musl-dev \
        gcc \
        python3 \
        python3-dev \
        bash \
        git \
        make \
        go \
        jq \
        zip \
        nodejs \
        npm

RUN python3 -m ensurepip --upgrade && pip3 install --upgrade pip
RUN pip3 install --upgrade awscli aws-sam-cli

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin /workdir
RUN wget -P /workdir https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-2.1.1-linux-x86_64.tar.bz2
RUN tar -xf /workdir/phantomjs-2.1.1-linux-x86_64.tar.bz2 -C /workdir
ADD deploy.sh /workdir/deploy.sh
RUN chmod +x /workdir/deploy.sh
CMD ["/workdir/deploy.sh"]
