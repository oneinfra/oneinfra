FROM ubuntu:18.04

RUN echo "Installing dependencies" \
  && apt-get update -qq \
  && apt-get install --no-install-recommends -qq -y \
       software-properties-common \
  && add-apt-repository ppa:wireguard/wireguard \
  && apt-get update -qq \
  && apt-get install -qq -y wireguard-dkms wireguard-tools \
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