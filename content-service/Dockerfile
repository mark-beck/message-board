FROM golang

WORKDIR /go/src/content-service

# Copy the local package files to the container's workspace.
COPY . .

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN ls .
RUN go install .

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/content-service

# Document that the service listens on port 8080.
EXPOSE 8080