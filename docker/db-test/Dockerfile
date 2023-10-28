ARG POSTGRES_VERSION
ARG MIGRATE_VERSION

# passing $MIGRATE_VERSION is impossible directly in COPY --from=...
FROM migrate/migrate:$MIGRATE_VERSION as migrate_image

FROM postgres:$POSTGRES_VERSION-alpine as builder

ARG PGTAP_VERSION

# system dependencies
RUN apk add --no-cache --update \
  git build-base make perl-dev clang15 llvm15

# install pg_prove
# WARNING: this (or next) step can break in many unexpected ways due to how awful cpan is.
# when this happens, rebuild with `--cache-from` option set to the previous layer.
RUN cpan TAP::Parser::SourceHandler::pgTAP

# install pgTAP
RUN git clone \
        --branch $PGTAP_VERSION \
        --depth 1 \
        https://github.com/theory/pgtap.git /pgtap \
  && cd /pgtap \
  && make \
  && make install

# TODO: final stage

COPY --from=migrate_image /usr/local/bin/migrate /bin/migrate

COPY non-root.sh root.sh /


ENTRYPOINT ["bash", "/non-root.sh"]
