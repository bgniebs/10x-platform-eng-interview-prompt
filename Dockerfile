FROM golang:1.6-alpine

# make app directory and add contents
RUN mkdir /app 
ADD app /app/ 
WORKDIR /app 

# copy csv file
COPY seattle-weather.csv /app/

# Set backend filename
ENV BACKEND_FILENAME="seattle-weather.csv"

# go build and start
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
CMD ["/app/main"]

# Expose port 3000
# TODO export port in ENV and update listen call in go
EXPOSE 3000