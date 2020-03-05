FROM ubuntu:19.10

RUN echo "Installing dependencies" \
  && apt-get update -qq \
  && apt-get install --no-install-recommends -qq -y haproxy \
  && mkdir -p /run/haproxy \
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

ENTRYPOINT ["/usr/sbin/haproxy", "-W", "-db", "-f", "/etc/haproxy/"]