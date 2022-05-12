FROM lfedge/eve-alpine:6.6.0 as build

ENV BUILD_PKGS git autoconf automake libtool gcc g++ openssl-dev build-base make socat gawk gnutls python3 libseccomp-dev
ENV PKGS libseccomp openssl
RUN eve-alpine-deploy.sh

# hadolint ignore=DL3018
RUN apk --no-cache --repository https://dl-cdn.alpinelinux.org/alpine/v3.13/main add -U libtasn1-dev expect gnutls
# hadolint ignore=DL3018
RUN apk --no-cache --repository https://dl-cdn.alpinelinux.org/alpine/edge/community add -U json-glib-dev

RUN git clone --branch=v0.9.4 --single-branch --depth=1 https://github.com/stefanberger/libtpms.git /libtpms
RUN cd /libtpms                                  &&\
    ./autogen.sh --prefix=/usr --libdir=/usr/lib --with-tpm2 --with-openssl &&\
    make -j "$(getconf _NPROCESSORS_ONLN)"                                  &&\
    make install && make DESTDIR=/out install


RUN git clone --branch=v0.7.3 --single-branch --depth=1 https://github.com/stefanberger/swtpm.git /swtpm
RUN cd /swtpm                                                &&\
    ./autogen.sh --prefix=/usr --libdir=/usr/lib --with-openssl --with-tss-user=tss --with-tss-group=tss &&\
    make -j "$(getconf _NPROCESSORS_ONLN)"                                                &&\
    make DESTDIR=/out install

FROM scratch
COPY --from=build /out /
ENTRYPOINT ["/usr/bin/swtpm"]
