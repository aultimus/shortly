FROM scratch
# Add pre-built executable
ADD shortly/shortly /go/bin/shortly
EXPOSE 8080
CMD ["/go/bin/shortly"]
