FROM snyk/snyk-cli:latest

WORKDIR /app

RUN apk add --no-cache git go

COPY . .

ENTRYPOINT ["snyk"]
CMD ["test", "--all-projects"]
