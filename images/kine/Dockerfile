FROM ubuntu:20.04 AS builder

ENV DEBIAN_FRONTEND noninteractive
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
  && GO111MODULE="on" go get -tags dqlite github.com/oneinfra/kine@801b3c3d52b33478991292497ee88ac4665b2c48 \
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


FROM ubuntu:20.04
COPY --from=builder /root/go/bin/kine /usr/local/bin/
RUN echo "Installing dependencies" \
  && apt-get update -qq \
  && apt-get install --no-install-recommends -qq -y \
       software-properties-common \
  && add-apt-repository -y ppa:dqlite/stable \
  && apt-get update -qq \
  && apt-get install --no-install-recommends -qq -y libdqlite0 libsqlite3-0 \
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