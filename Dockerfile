FROM alpine:latest

RUN apk -v --no-cache --update add bash jq npm git make musl-dev go gcc python3 python3-dev
RUN python3 -m ensurepip --upgrade && pip3 install --upgrade pip
RUN pip3 install --upgrade awscli aws-sam-cli
RUN pip3 uninstall --yes pip
RUN apk del python3-dev gcc musl-dev
RUN npm install -g babel-cli
RUN npm install -g babel-preset-minify

ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin
WORKDIR $GOPATH
ADD ./ /go/georgguessr
ENTRYPOINT ["tail", "-f", "/dev/null"]
