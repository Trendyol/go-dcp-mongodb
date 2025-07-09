FROM couchbase:7.6.3@sha256:dc0ca4b3b34fe9c063103539ed15ea4738e966c5f4978d004394aeb3ba99dd99

ADD configure.sh /configure.sh
RUN chmod +x /configure.sh
RUN echo " image starting"

EXPOSE 8091 8092 8093 8094 8095 8096 11207 11210 11211

CMD ["/configure.sh"] 