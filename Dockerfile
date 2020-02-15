FROM scratch
ADD main /
ADD config.json /
EXPOSE 9090
CMD ["/main"]
