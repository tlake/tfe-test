FROM alpine
RUN apk add --no-cache ca-certificates
ADD ./guestbook /
ADD ./views /views/
CMD ["/guestbook"]

