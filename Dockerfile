FROM golang AS build

# install dependencies
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash && \
    apt-get update && apt-get install -y nodejs

# copy and build the frontend and backend
COPY . /src
WORKDIR /src
RUN npm run-script --prefix frontend build && \
    go build -o app main.go



# Final Image
FROM debian

# copy compiled frontend and backend from build image
COPY --from=build /src/frontend/build /app/frontend/build
COPY --from=build /src/app /app/app

WORKDIR /app
ENTRYPOINT [ "./app" ]