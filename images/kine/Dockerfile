FROM ubuntu:18.04

ENV GOPATH /root/go
ENV PATH $PATH:/usr/local/go/bin:$GOPATH/bin

RUN echo "Installing dependencies" \
  && apt-get update -qq \
  && apt-get install --no-install-recommends -qq -y \
       software-properties-common wget git build-essential \
  && add-apt-repository -y ppa:dqlite/stable \
  && apt-get update -qq \
  && apt-get install --no-install-recommends -qq -y libdqlite-dev libsqlite3-dev \
  && wget https://dl.google.com/go/go1.12.16.linux-amd64.tar.gz -O go.tar.gz \
  && tar -xf go.tar.gz \
  && mv go /usr/local \
  && rm go.tar.gz \
  && GO111MODULE="on" go get -tags dqlite github.com/ereslibre/kine@6c55e1ae3b8864c3cbe95e6ae5e0d3c6e009f09b \
&& echo "Cleaning up image" \
  && apt-get clean -y \
  && rm -rf /var/cache/debconf/* \
            /var/lib/apt/lists/* \
            /var/log/* \
            /tmp/* \
            /var/tmp/* \
            /usr/share/doc/* \
            /usr/share/man/* \
            /usr/share/local/*


FROM ubuntu:18.04

COPY --from=0 /root/go/bin/kine /usr/local/bin/