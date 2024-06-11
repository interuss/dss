FROM kong:alpine
USER root
RUN mkdir /usr/local/share/lua/5.1/kong/plugins/scope-acl
COPY plugin/. /usr/local/share/lua/5.1/kong/plugins/scope-acl
USER kong